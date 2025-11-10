package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Response is the standard API response envelope
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

// ValidationErrorResponse includes field-specific errors
type ValidationErrorResponse struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error"`
	Fields    map[string]string `json:"fields"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp string            `json:"timestamp,omitempty"`
}

// Success writes a successful JSON response
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// SuccessWithRequest writes a successful response with request metadata
func SuccessWithRequest(w http.ResponseWriter, statusCode int, data interface{}, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:   true,
		Data:      data,
		RequestID: extractRequestID(r),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:   false,
		Error:     message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// ValidationError writes a 400 response with field errors
func ValidationError(w http.ResponseWriter, fieldErrors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ValidationErrorResponse{
		Success:   false,
		Error:     "validation failed",
		Fields:    fieldErrors,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// extractRequestID gets or generates a request ID
func extractRequestID(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return requestID
}
