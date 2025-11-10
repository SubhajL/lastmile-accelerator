package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

var (
	ErrTooLarge    = errors.New("payload too large")
	ErrInvalidJSON = errors.New("invalid json")
	ErrUnknownField = errors.New("unknown field")
)

const (
	DefaultMaxBody     int64 = 1 << 20 // 1 MiB
	ErrorIngestMaxBody int64 = 2 << 20 // 2 MiB
)

// StrictDecode enforces a maximum size and disallows unknown fields.
// It wraps the request body with http.MaxBytesReader, decodes exactly one JSON value,
// and returns typed errors that callers can map to HTTP statuses.
func StrictDecode(w http.ResponseWriter, r *http.Request, v any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBody
	}
	// Enforce body size limit
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		// MaxBytesReader surfaces as *http.MaxBytesError
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return ErrTooLarge
		}
		// Some Go versions return a plain error string for oversized bodies
		if strings.Contains(err.Error(), "request body too large") {
			return ErrTooLarge
		}
		// Unknown fields are reported via error text
		if isUnknownFieldError(err) {
			return ErrInvalidJSON
		}
		// Syntax/type errors
		switch err := err.(type) {
		case *json.SyntaxError, *json.UnmarshalTypeError:
			_ = err
			return ErrInvalidJSON
		default:
			if errors.Is(err, io.EOF) {
				return ErrInvalidJSON
			}
			return ErrInvalidJSON
		}
	}
	// Ensure there's no trailing data
	if dec.More() {
		return ErrInvalidJSON
	}
	return nil
}

func isUnknownFieldError(err error) bool {
	// json.Decoder returns errors like: "json: unknown field \"foo\""
	return strings.Contains(err.Error(), "unknown field")
}
