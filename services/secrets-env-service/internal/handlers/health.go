package handlers

import (
	"net/http"
)

// HealthCheck returns a simple OK response for liveness probes
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
}
