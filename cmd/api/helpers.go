package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mostafejur21/greenlight_go/internal/validator"
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
	// set the max json body size.
	maxBytes := 1_048_576 // (1MB)
	// use the r.MaxBytesReader () to limit the max read
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	// this DisallowUnknownFields () methods will not allow any unknown fields.
	dec.DisallowUnknownFields()
	// decode the json body into the terget destination (dst)
	err := dec.Decode(dst)
	if err != nil {
		// if there is any error during the decoding phase
		var syntextError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError
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
		// if the json has a field that cannot turn into the destination.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fildName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contain unknown fields %s", fildName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		// this json.InvalidUnmarshalError error will be returned if we pass something that is not
		// a non-nil pointer to Decode(), we catch that error and panic and stop the server
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		// for anything else, we just simply return the err
		default:
			return err
		}
	}
	// calling the Decode() again using a anonymous struct to check that there is any second JSON in the body. cause
	// the body should contain only single json body
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body only contain a single JSON value")
	}
	return nil
}

// The readString() helper method will returns a string value from query string, or the provided
// default value if no matching key could not found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exist, this will return the empty string"".
	s := qs.Get(key)

	// if no key exists (or the value is empty) then return the default value.
	if s == "" {
		return defaultValue
	}

	// otherwise we will return the S
	return s
}

// the readCSV() helper method reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns the
// provided default value.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	// this will parse the value into a []string slice and return it
	return strings.Split(csv, ",")
}

// The readInt() helper method reads a string value from the query string and converts it to an
// integer before returning it. if no matching value found, then it will return the default value.
// if the value could not be convert into an integer, then we record an error message in the provided
// validator instance
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	// try to convert the string value into an integer value, if this failed, then add an error message
	// validator instance
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddErrors(key, "must be an integer value")
		return defaultValue
	}

	return i

}

// the background() helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	// launch a backgroun goroutine
	go func() {
		// Recover any panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		//Execute the arbitrary function that we passed as the parameter
		fn()
	}()
}
