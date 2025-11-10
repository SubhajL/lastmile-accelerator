package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// sensitiveKeys lists field names that should be redacted
var sensitiveKeys = map[string]bool{
	"password":   true,
	"secret":     true,
	"api_key":    true,
	"token":      true,
	"secret_key": true,
	"private_key": true,
	"access_key": true,
	"credential": true,
}

// New creates a new zerolog logger with JSON output
func New(serviceName, logLevel string, writer io.Writer) zerolog.Logger {
	if writer == nil {
		writer = os.Stdout
	}

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// WithCorrelationID adds a correlation ID to the logger context
func WithCorrelationID(logger zerolog.Logger, correlationID string) zerolog.Logger {
	return logger.With().Str("correlation_id", correlationID).Logger()
}

// WithTenantContext adds tenant and project IDs to the logger context
func WithTenantContext(logger zerolog.Logger, tenantID, projectID string) zerolog.Logger {
	return logger.With().
		Str("tenant_id", tenantID).
		Str("project_id", projectID).
		Logger()
}

// RedactSecrets sanitizes a map by replacing sensitive values with redaction markers
func RedactSecrets(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range data {
		lowerKey := strings.ToLower(key)
		
		// Check if this key should be redacted
		shouldRedact := false
		for sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				shouldRedact = true
				break
			}
		}

		if shouldRedact {
			result[key] = "***REDACTED***"
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively redact nested maps
			result[key] = RedactSecrets(nestedMap)
		} else {
			result[key] = value
		}
	}
	
	return result
}

// SafeLogError logs an error without exposing sensitive details
func SafeLogError(logger zerolog.Logger, err error, msg string) {
	if err == nil {
		return
	}

	// Log error type but sanitize message
	errMsg := err.Error()
	
	// Remove common secret patterns from error messages
	secretPatterns := []string{
		"sk-",
		"ghp_",
		"gho_",
		"Bearer ",
		"token=",
		"password=",
	}
	
	for _, pattern := range secretPatterns {
		if strings.Contains(errMsg, pattern) {
			errMsg = "[error message contains sensitive data]"
			break
		}
	}
	
	logger.Error().Str("error", errMsg).Msg(msg)
}
