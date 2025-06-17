package main

import (
	"net/http"
)

// healthcheckHandler provides a simple health check endpoint that returns the application status
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create health status response
	health := envelope{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
		"database":    "okay",
	}

	// Write successful health check response
	err := app.writeJSON(w, http.StatusOK, health, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
