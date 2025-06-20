package main

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestReadString(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "parameter exists",
			queryParams:  "limit=10&country=us",
			key:          "limit",
			defaultValue: "20",
			expected:     "10",
		},
		{
			name:         "parameter doesn't exist",
			queryParams:  "country=us",
			key:          "limit",
			defaultValue: "20",
			expected:     "20",
		},
		{
			name:         "empty parameter value",
			queryParams:  "limit=&country=us",
			key:          "limit",
			defaultValue: "20",
			expected:     "20",
		},
		{
			name:         "no query parameters",
			queryParams:  "",
			key:          "limit",
			defaultValue: "20",
			expected:     "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with query parameters
			req := httptest.NewRequest("GET", "/?"+tt.queryParams, nil)
			
			// Create test app
			app := &application{}
			
			result := app.readString(req.URL.Query(), tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("readString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildAPIURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		params   map[string]string
		expected string
		hasError bool
	}{
		{
			name:     "basic URL with no params",
			baseURL:  "https://api.example.com",
			endpoint: "/search",
			params:   map[string]string{},
			expected: "https://api.example.com/search",
			hasError: false,
		},
		{
			name:     "URL with params",
			baseURL:  "https://api.example.com/v2",
			endpoint: "/everything",
			params: map[string]string{
				"q":       "music",
				"apiKey":  "test123",
				"country": "us",
			},
			expected: "https://api.example.com/v2/everything?apiKey=test123&country=us&q=music",
			hasError: false,
		},
		{
			name:     "URL with empty params filtered out",
			baseURL:  "https://api.example.com",
			endpoint: "/search",
			params: map[string]string{
				"q":     "music",
				"empty": "",
				"limit": "10",
			},
			expected: "https://api.example.com/search?limit=10&q=music",
			hasError: false,
		},
		{
			name:     "trailing slash handling",
			baseURL:  "https://api.example.com/",
			endpoint: "/search",
			params: map[string]string{
				"q": "music",
			},
			expected: "https://api.example.com/search?q=music",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildAPIURL(tt.baseURL, tt.endpoint, tt.params)
			
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			
			if !tt.hasError {
				// Parse both URLs to compare query params (order may vary)
				expectedURL, _ := url.Parse(tt.expected)
				resultURL, _ := url.Parse(result)
				
				if expectedURL.Scheme != resultURL.Scheme ||
					expectedURL.Host != resultURL.Host ||
					expectedURL.Path != resultURL.Path {
					t.Errorf("buildAPIURL() base parts = %v, want %v", result, tt.expected)
				}
				
				// Check that all expected params are present
				expectedParams := expectedURL.Query()
				resultParams := resultURL.Query()
				
				for key, values := range expectedParams {
					if resultParams.Get(key) != values[0] {
						t.Errorf("buildAPIURL() missing or incorrect param %s = %v, want %v", 
							key, resultParams.Get(key), values[0])
					}
				}
			}
		})
	}
}
