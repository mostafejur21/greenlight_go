package main

import (
	"fmt"
	"net/http"
)

// the logError() method is a generic helper function for logging any error message
// along with the passing url
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// the errorResponse() method is a helper function for sending JSON-Formatted error response
// to the client with a appropate status code
// here 500 status means internal server error
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// the serverErrorResponse() method will be used when our application encounterd any unexpected
// problem at runtime (server side error - 500). then we will send a response with the help of
// the errorResponse() method to send a 500 error with a error message (JSON-Formatted)
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not be process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// the notFoundResponse() method will be used to send the 404 not found error and a json message to the user
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resources could not found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// the methodNotAllowedResponse () method will be send the 405 status error message to the user if user
// enter the wrong request, like GET instead of POST
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resources", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestRespons(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Note that here the error is a map of string, string. same as the errors map in our Validation type.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate Limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	msg := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}
