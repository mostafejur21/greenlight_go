package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template/parse"
)

// Define an error if the UnmarshalJSON () method is unable to parse or convert the JSON to Runtime
var ErrInvalidRunTimeFormat = errors.New("invalid runtime format")
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	// generate a string containing Movie runtime with required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// this will wrap the jsonValue in a double qoute "".
	// strconv.Quote() will wrap the jsonValue inside a "". otherwise the json will give a error
	qoutedJSON := strconv.Quote(jsonValue)

	return []byte(qoutedJSON), nil
}

// implementing a UnmarshalJSON() method on the Runtime type so that it satisfies the
// json.UnmarshalJSON interface. IMPORTANT NOTE: because json.UnmarshalJSON needs to modify
// the receiver value, we must pass the pointer receiver for this to work correctly
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
    // we expect that the incomming runtime value will be a string type. "<runtime> mins".
    // so first thing to do is remove the double qoute. if we cannot unqoute, then we can
    // throw the ErrInvalidRunTimeFormat error
    unquoteJSONValue, err := strconv.Unquote(string(jsonValue))
    if err != nil {
        return ErrInvalidRunTimeFormat
    }
    // split the value to isolate the part that contain the number
    parts := strings.Split(unquoteJSONValue, " ")

    // checking the parts that it is the correct value, otherwise throw the ErrInvalidRunTimeFormat
    if len(parts) != 2 || parts[1] != "mins" {
        return ErrInvalidRunTimeFormat
    }

    // now parse that number into a int32
    i, err := strconv.ParseInt(parts[0], 10, 32)
    if err != nil {
        return ErrInvalidRunTimeFormat
    }

    // now convert that int32 into a Runtime Type and assign it to the receiver. Note that we use the
    // * operator to set the underlying value pointer
    *r = Runtime(i)
    return nil
}
