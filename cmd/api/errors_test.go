package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestErrorResponses(t *testing.T) {
	app := &application{
		logger: zap.NewNop(),
	}

	tests := []struct {
		name           string
		errorFunc      func(http.ResponseWriter, *http.Request)
		expectedStatus int
		expectedField  string
	}{
		{
			name: "bad request error",
			errorFunc: func(w http.ResponseWriter, r *http.Request) {
				app.badRequestResponse(w, r, errors.New("invalid parameter"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedField:  "error",
		},
		{
			name: "not found error",
			errorFunc: func(w http.ResponseWriter, r *http.Request) {
				app.notFoundResponse(w, r)
			},
			expectedStatus: http.StatusNotFound,
			expectedField:  "error",
		},
		{
			name: "invalid authentication token",
			errorFunc: func(w http.ResponseWriter, r *http.Request) {
				app.invalidAuthenticationTokenResponse(w, r)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedField:  "error",
		},
		{
			name: "server error",
			errorFunc: func(w http.ResponseWriter, r *http.Request) {
				app.serverErrorResponse(w, r, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedField:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request and response recorder
			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()

			// Call the error function
			tt.errorFunc(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check content type
			expectedContentType := "application/json"
			if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
				t.Errorf("expected content type %s, got %s", expectedContentType, contentType)
			}

			// Check response body structure
			var response map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			// Check that error field exists
			if _, exists := response[tt.expectedField]; !exists {
				t.Errorf("expected field %s not found in response", tt.expectedField)
			}
		})
	}
}

func TestValidationErrorResponse(t *testing.T) {
	app := &application{
		logger: zap.NewNop(),
	}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Create some validation errors
	errors := map[string]string{
		"email":    "must be a valid email address",
		"password": "must be at least 8 characters long",
	}

	app.failedValidationResponse(rr, req, errors)

	// Check status code
	expectedStatus := http.StatusUnprocessableEntity
	if rr.Code != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, rr.Code)
	}

	// Check response structure
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	// Check that errors field exists
	if _, exists := response["error"]; !exists {
		t.Errorf("expected 'error' field not found in response")
	}
}
