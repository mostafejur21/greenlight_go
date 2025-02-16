package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mostafejur21/greenlight_go/internal/data"
	"github.com/mostafejur21/greenlight_go/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"Password"`
	}

	// Parse the request body into the struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestRespons(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the Password.Set() method to generate and store the hashed and plaintext password
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If error is ErrDuplicateEmail
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddErrors("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Add the "movies:read" permission to the new user
	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// generate token
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// background go routine for sending the email
	//	go func() {
	//        // running a deferred function which uses recover() to catch any panic, and log an
	//        // error message instead of terminating the application
	//        defer func() {
	//            if err := recover();  err != nil {
	//                app.logger.Error(fmt.Sprintf("%v", err))
	//            }
	//        }()
	//		// send the mail to the user's email using SMTP mail
	//		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
	//
	//		if err != nil {
	//			// if there is an error sending the email then we use the
	//			// app.logger.Error() helper to manage it
	//			app.logger.Error(err.Error())
	//		}
	//
	//	}()
	// Using the app.background() helper function to replace the background goroutine
	app.background(func() {
		// now, there are multiple data that we want to pass to our email templates (user id, token)
		// so creating a map and 'holding structure' for the data
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userId":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})
	// send an json response to the client
	// change code 202 instead of 201, this StatusAccepted indecates that the request has been accepted
	// for processing, but the processing has not been completed because of the background go routine
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext token from the request body
	var input struct {
		TokenPlainText string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestRespons(w, r, err)
		return
	}

	// Validate the plaintext token provided by the user
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlainText); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrive the details of the user from the token using GetForToken() method. If no matching, then
	// let the client know that the token is not valid
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlainText)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddErrors("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user activation status
	user.Activated = true

	// Save the updated user record in our db, checking for any edit conflicts
	err = app.models.Users.Update(user)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Deleting all activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
