package main

import (
	"errors"
	"fmt"
	"net/http"
	"github.com/mostafejur21/greenlight_go/internal/data"
	"github.com/mostafejur21/greenlight_go/internal/validator"
)

const must_provided string = "must be provided"

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestRespons(w, r, err)
		return
	}

	// Copy the values from the input to our own Movies struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		RunTime: input.Runtime,
		Genres:  input.Genres,
	}
	// initialize a new Validator instance
	v := validator.New()

	// use the v.valid() method to see if any check failed. if they did, then use the
	// call the ValidateMovie() function and return a response containig the errors if any
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// here calling the Insert() method on our movies model, passing in a pointer to the
	// validated movie struct. this will create a record in the database and update the
	// movies struct with the system generated information
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// when sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find that movie.
	// custom header
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// send a json response body with a status 201 (created),
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		// custom not found error method
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// here using custom erro method for showing the server error response
		app.serverErrorResponse(w, r, err)
	}

}
