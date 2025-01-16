package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	// generate a string containing Movie runtime with required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// this will wrap the jsonValue in a double qoute "".
	// strconv.Quote() will wrap the jsonValue inside a "". otherwise the json will give a error
	qoutedJSON := strconv.Quote(jsonValue)

	return []byte(qoutedJSON), nil
}
