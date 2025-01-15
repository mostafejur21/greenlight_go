package main

import (
	"encoding/json"
	"errors"
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
