package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// define an envelope type for enveloping the json response
type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	// when httprouter parsing a request, any interpolated URL parameters will be
	// stored in the request context. we can use the ParamsFromContext() function
	// to retrive a slice containing this parameters names and values
	params := httprouter.ParamsFromContext(r.Context())

	// use the ByName() methods to get the id.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id params")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data into JSON, return error if there is any
	// Using the json.MarshalIndent () instead of the json.Marshal () so that extra whitespace is added
	// to the encoding JSON. here we use no line prefix "", and tab index ("\t"), for each element

	// But using the json.MarshalIndent() function instade of json.Marshal() will impact on the performance slightly.
	// Because go need's to do extra work for the whitespace and the response []byte will be slightly larger. but not that big of a deal
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// append a new line to make it easy to read it in the terminal
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	// Adding the w.headers ("Content-type", "application/json")
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// decode the json body into the terget destination (dst)
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		// if there is any error during the decoding phase
		var syntextError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// use the errors.As() function to check weather the error has a type of *json.SyntaxError
		case errors.As(err, &syntextError):
			return fmt.Errorf("body contains badly formed JSON (at the %d)", syntextError.Offset)
		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contain badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("stop contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains badly formed JSON (at the %d)", syntextError.Offset)
        case errors.Is(err, io.EOF):
            return errors.New("body must not be empty")
        // this json.InvalidUnmarshalError error will be returned if we pass something that is not
        // a non-nil pointer to Decode(), we catch that error and panic and stop the server
        case errors.As(err, &invalidUnmarshalError):
            panic(err)
        // for anything else, we just simply return the err
        default:
            return err
		}
	}
    return nil
}
