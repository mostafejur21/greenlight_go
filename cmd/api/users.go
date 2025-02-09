package main

import (
	"errors"
	"net/http"

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
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
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
