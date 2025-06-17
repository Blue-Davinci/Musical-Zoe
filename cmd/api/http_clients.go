package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Client represents the central HTTP client with retry capabilities
type Optivet_Client struct {
	httpClient *retryablehttp.Client
}

// NewClient initializes and returns a new Client with custom configurations
func NewClient(timeout time.Duration, retries int) *Optivet_Client {
	// Create a retryable HTTP client with custom settings
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retries
	retryClient.HTTPClient.Timeout = timeout
	// Use a much faster backoff strategy for better user experience
	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		// Simple fixed backoff of 1 second between retries
		return 1 * time.Second
	}
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.Logger = nil

	return &Optivet_Client{
		httpClient: retryClient,
	}
}

// GETRequest sends a GET request to the specified URL and unmarshals the response into a generic type T
func GETRequest[T any](c *Optivet_Client, url string, headers map[string]string) (T, error) {
	var result T

	// Create a new request
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Check if the response status is not 2xx
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := fmt.Sprintf("non-2xx response code: %d | url: %s", resp.StatusCode, url)
		return result, errors.New(message)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	// Unmarshal the response into the provided generic type
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}
	//fmt.Printf("Response: %v", result)

	return result, nil
}

// GETRequestWithParams sends a GET request with query parameters and unmarshals the response into a generic type T
func GETRequestWithParams[T any](c *Optivet_Client, baseURL string, params map[string]string, headers map[string]string) (T, error) {
	var result T

	// Parse the base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return result, err
	}

	// Add query parameters
	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	// Create a new request
	req, err := retryablehttp.NewRequest("GET", u.String(), nil)
	if err != nil {
		return result, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Check if the response status is not 2xx
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := fmt.Sprintf("non-2xx response code: %d | url: %s", resp.StatusCode, u.String())
		return result, errors.New(message)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	// Unmarshal the response into the provided generic type
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
