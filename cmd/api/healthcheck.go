package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// js := `{"status": "available", "environment": %q, "version": %q}`
	// now create a map that will hold the apps information
    // here call the envelope type for envelope the json response
	envelope := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	// By passing the data map to json.Marshal() function, this json.Marshal() function
	// will return a []byte (slice of byte) and error.
	// Why do we need Marshal() => to encode our GO code into JSON format
	err := app.writeJSON(w, http.StatusOK, envelope, nil)
	if err != nil {
        // here using custom erro method for showing the server error response
        app.serverErrorResponse(w, r, err)
	}
}
