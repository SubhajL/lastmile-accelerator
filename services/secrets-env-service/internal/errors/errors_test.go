package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_ErrorMessage(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		Message:    "validation failed",
		Internal:   errors.New("internal details"),
	}

	// Error() should return user-safe message only
	assert.Equal(t, "validation failed", apiErr.Error())
}

func TestAPIError_StatusCode(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 404,
		Message:    "not found",
	}

	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestNewAPIError_WrapsInternalError(t *testing.T) {
	internal := errors.New("database connection failed")
	apiErr := NewAPIError(500, "service unavailable", internal)

	assert.Equal(t, 500, apiErr.StatusCode)
	assert.Equal(t, "service unavailable", apiErr.Message)
	
	// Should be able to unwrap to get internal error
	unwrapped := errors.Unwrap(apiErr)
	assert.Equal(t, internal, unwrapped)
}

func TestIsRetryable_NetworkTimeout(t *testing.T) {
	err := fmt.Errorf("dial tcp: i/o timeout")
	assert.True(t, IsRetryable(err))
}

func TestIsRetryable_503Status(t *testing.T) {
	apiErr := NewAPIError(503, "service unavailable", nil)
	assert.True(t, IsRetryable(apiErr))
}

func TestIsRetryable_ValidationError(t *testing.T) {
	apiErr := NewAPIError(400, "bad request", nil)
	assert.False(t, IsRetryable(apiErr))
}

func TestHandlePanic_RecoverAndLog(t *testing.T) {
	recovered := false
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		
		defer HandlePanic()
		panic("test panic")
	}()

	// HandlePanic should have recovered, so outer recover shouldn't trigger
	assert.False(t, recovered)
}

func TestHandlePanic_LogsStackTrace(t *testing.T) {
	// This test verifies HandlePanic doesn't crash
	// In real usage, it would log to a logger
	func() {
		defer HandlePanic()
		panic("test panic with stack")
	}()
	
	// If we get here, HandlePanic worked
	assert.True(t, true)
}

func TestSentinelErrors(t *testing.T) {
	assert.Equal(t, "secret not found", ErrSecretNotFound.Error())
	assert.Equal(t, "unauthorized access", ErrUnauthorized.Error())
	assert.Equal(t, "validation failed", ErrValidation.Error())
}
