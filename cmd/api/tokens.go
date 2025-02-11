package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mostafejur21/greenlight_go/internal/data"
	"github.com/mostafejur21/greenlight_go/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse the email and passwork from the req body
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestRespons(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	// check if the provided password matches the actual password for the user
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// if the password don't match
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// if the password match, then we generate the new token with 24 hour expiry time
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Encode the token to JSON and send it in the response along with a 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication-token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
