package errors

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
)

// Sentinel errors
var (
	ErrSecretNotFound = errors.New("secret not found")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrValidation     = errors.New("validation failed")
)

// APIError represents an HTTP-aware error
type APIError struct {
	StatusCode int
	Message    string
	Internal   error
	RequestID  string
}

// Error returns the user-safe message
func (e *APIError) Error() string {
	return e.Message
}

// Unwrap returns the internal error for errors.Unwrap
func (e *APIError) Unwrap() error {
	return e.Internal
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, message string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Internal:   err,
	}
}

// IsRetryable determines if an error is transient and worth retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for APIError with retryable status codes
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 503 || 
		       apiErr.StatusCode == 502 || 
		       apiErr.StatusCode == 504 ||
		       apiErr.StatusCode == 429
	}

	// Check for common network error patterns
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"too many open files",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// HandlePanic recovers from panics and logs the stack trace
func HandlePanic() {
	if r := recover(); r != nil {
		// In production, this would use the logger
		// For now, just recover and print to stderr
		fmt.Printf("PANIC: %v\nStack:\n%s\n", r, debug.Stack())
	}
}
