package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mostafejur21/greenlight_go/internal/data"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "creating a new movie")
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
