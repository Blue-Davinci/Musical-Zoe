package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Blue-Davinci/musical-zoe/internal/data"
	"go.uber.org/zap"
)

func TestHealthcheckHandler(t *testing.T) {
	// Create a test application with minimal setup
	app := &application{
		config: config{
			env: "test",
		},
		logger: zap.NewNop(),  // No-op logger for tests
		models: data.Models{}, // Empty models for this test
	}

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler directly
	app.healthcheckHandler(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body structure
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	// Check required fields exist
	expectedFields := []string{"status", "environment", "version"}
	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("expected field %s not found in response", field)
		}
	}

	// Check specific values
	if response["status"] != "available" {
		t.Errorf("expected status 'available', got %v", response["status"])
	}

	if response["environment"] != "test" {
		t.Errorf("expected environment 'test', got %v", response["environment"])
	}
}
