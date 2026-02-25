package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"example.com/lma/secrets-env-service/internal/startup"
)

type readyzResponse struct {
	Ready  bool              `json:"ready"`
	Checks map[string]string `json:"checks"`
}

func ReadyCheck(readiness startup.Readiness) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		rep, ready := readiness.Check(ctx)
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(readyzResponse{Ready: ready, Checks: rep.Checks})
	}
}
