package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
    // Encode the data into JSON, return error if there is any
    js, err := json.Marshal(data)
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
