package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccess_ReturnsJSONWithData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	Success(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Empty(t, response.Error)
}

func TestError_ReturnsUserSafeMessage(t *testing.T) {
	w := httptest.NewRecorder()
	internalErr := errors.New("database connection failed")

	Error(w, http.StatusInternalServerError, "service unavailable", internalErr)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "service unavailable", response.Error)
	// Internal error should NOT be in response
	assert.NotContains(t, response.Error, "database connection")
}

func TestValidationError_Returns400(t *testing.T) {
	w := httptest.NewRecorder()
	fieldErrors := map[string]string{
		"key":         "key is required",
		"environment": "must be dev, staging, or prod",
	}

	ValidationError(w, fieldErrors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ValidationErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "validation failed", response.Error)
	assert.Len(t, response.Fields, 2)
	assert.Equal(t, "key is required", response.Fields["key"])
}

func TestExtractRequestID_FromHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-req-123")

	requestID := extractRequestID(req)

	assert.Equal(t, "test-req-123", requestID)
}

func TestExtractRequestID_Generates(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	requestID := extractRequestID(req)

	assert.NotEmpty(t, requestID)
	assert.Len(t, requestID, 36) // UUID format
}

func TestResponse_IncludesMetadata(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "req-456")

	SuccessWithRequest(w, http.StatusOK, map[string]string{"result": "ok"}, req)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "req-456", response.RequestID)
	assert.NotEmpty(t, response.Timestamp)
}
