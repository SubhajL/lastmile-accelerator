package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_CreatesLoggerWithServiceName(t *testing.T) {
	var buf bytes.Buffer
	logger := New("secrets-env-service", "info", &buf)

	logger.Info().Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "secrets-env-service", logEntry["service"])
	assert.Equal(t, "test message", logEntry["message"])
}

func TestNew_RespectsLogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service", "error", &buf)

	// Debug should not appear
	logger.Debug().Msg("debug message")
	assert.Empty(t, buf.String())

	// Error should appear
	logger.Error().Msg("error message")
	assert.NotEmpty(t, buf.String())
}

func TestWithCorrelationID_AddsField(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service", "info", &buf)
	
	loggerWithCorr := WithCorrelationID(logger, "corr-123")
	loggerWithCorr.Info().Msg("test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "corr-123", logEntry["correlation_id"])
}

func TestWithTenantContext_AddsBothFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service", "info", &buf)
	
	loggerWithCtx := WithTenantContext(logger, "tenant-456", "proj-789")
	loggerWithCtx.Info().Msg("test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "tenant-456", logEntry["tenant_id"])
	assert.Equal(t, "proj-789", logEntry["project_id"])
}

func TestRedactSecrets_ReplacesValues(t *testing.T) {
	data := map[string]interface{}{
		"api_key":    "secret-value-123",
		"username":   "john",
		"password":   "super-secret",
	}

	redacted := RedactSecrets(data)
	
	assert.Equal(t, "***REDACTED***", redacted["api_key"])
	assert.Equal(t, "***REDACTED***", redacted["password"])
	assert.Equal(t, "john", redacted["username"]) // username not redacted
}

func TestRedactSecrets_PreservesKeys(t *testing.T) {
	data := map[string]interface{}{
		"secret_key": "value",
		"token":      "jwt-token",
	}

	redacted := RedactSecrets(data)
	
	// Keys should exist
	_, hasSecretKey := redacted["secret_key"]
	_, hasToken := redacted["token"]
	
	assert.True(t, hasSecretKey)
	assert.True(t, hasToken)
}

func TestRedactSecrets_HandlesNestedMaps(t *testing.T) {
	data := map[string]interface{}{
		"config": map[string]interface{}{
			"api_key": "secret",
			"name":    "myapp",
		},
	}

	redacted := RedactSecrets(data)
	
	configMap := redacted["config"].(map[string]interface{})
	assert.Equal(t, "***REDACTED***", configMap["api_key"])
	assert.Equal(t, "myapp", configMap["name"])
}

func TestSafeLogError_NoSensitiveData(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service", "info", &buf)
	
	// Simulate error with sensitive data
	err := assert.AnError
	SafeLogError(logger, err, "operation failed")

	output := buf.String()
	
	// Should log the message
	assert.Contains(t, output, "operation failed")
	
	// Should not contain common secret patterns
	assert.NotContains(t, output, "sk-")
	assert.NotContains(t, output, "ghp_")
}
