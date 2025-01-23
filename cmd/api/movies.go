package main

import (
	"errors"
	"fmt"
	"github.com/mostafejur21/greenlight_go/internal/data"
	"github.com/mostafejur21/greenlight_go/internal/validator"
	"net/http"
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
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
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

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the movie id from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database using the id. sending the 404 Not Found
	// response to the client if we cannot found the movie
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidRunTimeFormat):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Declare an input struct to hold the expected data from the client,
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read the JSON request body data into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestRespons(w, r, err)
		return
	}

	// Copy the value from the request body to the appropriate fields of the movie record
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.RunTime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// validate the update movie record
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the update movies in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movies from the DB, sending a 404 not found response to the client if there is no matching
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully dleleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// Initialize a new validator instance.
	v := validator.New()

	// Call the r.Url.Query() to get the url.values map containig the query string data.
	qs := r.URL.Query()

	// using our readCSV/readInt/readString helper function to extract the title and genres query string value
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// reading the Filters
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to id if it not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// adding the sort safelist value for checking
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

    // Execute the validaton check on the Filters struct and send a response
	// check the validator instance for any errors and use the failedValidationResponse()
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
