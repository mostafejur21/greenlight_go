package main

import (
	"fmt"
	"net/http"
	"time"

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
        Title: input.Title,
        Year: input.Year,
        RunTime: input.Runtime,
        Genres: input.Genres,
    }
	// initialize a new Validator instance
	v := validator.New()

    // use the v.valid() method to see if any check failed. if they did, then use the
    // call the ValidateMovie() function and return a response containig the errors if any
    if data.ValidateMovie(v, movie); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }
	fmt.Fprintf(w, "%+v", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		// custom not found error method
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "BatMan",
		RunTime:   102,
		Genres:    []string{"drama", "action", "war"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// here using custom erro method for showing the server error response
		app.serverErrorResponse(w, r, err)
	}

}
