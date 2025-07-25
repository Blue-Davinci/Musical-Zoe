package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Blue-Davinci/musical-zoe/internal/data"
	"github.com/Blue-Davinci/musical-zoe/internal/validator"
)

var (
	ErrInvalidAuthentication = errors.New("invalid authentication token format")
	ErrNoDataFoundInRedis    = errors.New("no data found in Redis")
)

// Define an envelope type.
type envelope map[string]any

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')
	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include.
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Add the "Content-Type: application/json" header, then write the status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	// Decode the request body to the destination.
	err := dec.Decode(dst)
	err = app.jsonReadAndHandleError(err)
	if err != nil {
		return err
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// jsonReadAndHandleError() is a helper function that takes an error as a parameter and
// returns a cleaned-up error message. This is used to provide more information in the
// event of a JSON decoding error.
func (app *application) jsonReadAndHandleError(err error) error {
	if err != nil {
		// Vars to carry our errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		// Add a new maxBytesError variable.
		var maxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"". We check for this, extract the field name from the error,
		// and interpolate it into our custom error message. Note that there's an open
		// issue at https://github.com/golang/go/issues/29035 regarding turning this
		// into a distinct error type in the future.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// Use the errors.As() function to check whether the error has the type
		// *http.MaxBytesError. If it does, then it means the request body exceeded our
		// size limit of 1MB and we return a clear error message.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	return nil
}

// The background() helper accepts an arbitrary function as a parameter.
// It launches a background goroutine to execute the function.
// The done() method of the WaitGroup is called when the goroutine completes.
func (app *application) background(fn func()) {
	app.wg.Add(1)
	// Launch a background goroutine.
	go func() {
		//defer our done()
		defer app.wg.Done()
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%s", err))
			}
		}()
		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}

// aunthenticatorHelper() is a helper function for the authentication middleware
// It takes in a request and returns a user and an error
func (app *application) aunthenticatorHelper(r *http.Request) (*data.User, error) {
	// Retrieve the value of the Authorization header from the request. This will
	authorizationHeader := r.Header.Get("Authorization")
	// If there is no Authorization header found, use the contextSetUser() helper to
	// add the AnonUser to the request context. Then we
	if authorizationHeader == "" {
		return data.AnonymousUser, nil
	}
	// Otherwise, we expect the value of the Authorization header to be in the format
	// "Bearer <token>". We try to split this into its constituent parts, and if the
	// header isn't in the expected format we return a 401 Unauthorized response
	// using the invalidAuthenticationTokenResponse() helper
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, ErrInvalidAuthentication
	}
	// Extract the actual authentication token from the header parts.
	token := headerParts[1]
	//app.logger.Info("User id Connected", zap.String("Connected ID", token))
	// Validate the token to make sure it is in a sensible format.
	v := validator.New()
	// If the token isn't valid, use the invalidAuthenticationTokenResponse()
	// helper to send a response, rather than the failedValidationResponse() helper
	// that we'd normally use.
	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		return nil, ErrInvalidAuthentication
	}
	// Retrieve the details of the user associated with the authentication token,
	// again calling the invalidAuthenticationTokenResponse() helper if no
	// matching record was found. IMPORTANT: Notice that we are using
	// ScopeAuthentication as the first parameter here.
	user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrGeneralRecordNotFound):
			return nil, ErrInvalidAuthentication
		default:
			return nil, ErrInvalidAuthentication
		}
	}
	return user, nil
}

// The readString() helper returns a string value from the query string, or the provided
// default value if no matching key could be found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return the empty string "".
	s := qs.Get(key)
	// If no key exists (or the value is empty) then return the default value.
	if s == "" {
		return defaultValue
	}
	// Otherwise return the string.
	return s
}

// buildAPIURL constructs a full API URL with query parameters
func buildAPIURL(baseURL, endpoint string, params map[string]string) (string, error) {
	// Parse the base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Add the endpoint to the path
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(endpoint, "/")

	// Add query parameters
	q := u.Query()
	for key, value := range params {
		if value != "" { // Only add non-empty parameters
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
