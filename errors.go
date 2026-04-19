package sleeper

import (
	"errors"
	"fmt"
)

// Sentinel errors for common API failure modes.
var (
	ErrNotFound    = errors.New("sleeper: resource not found")
	ErrRateLimited = errors.New("sleeper: rate limited")
	ErrServerError = errors.New("sleeper: server error")
)

// APIError represents a non-2xx response from the Sleeper API.
type APIError struct {
	StatusCode int
	Endpoint   string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("sleeper: %d %s: %s", e.StatusCode, e.Endpoint, e.Body)
}

// Unwrap returns the sentinel error matching the status code, if any.
func (e *APIError) Unwrap() error {
	switch {
	case e.StatusCode == 404:
		return ErrNotFound
	case e.StatusCode == 429:
		return ErrRateLimited
	case e.StatusCode >= 500:
		return ErrServerError
	default:
		return nil
	}
}
