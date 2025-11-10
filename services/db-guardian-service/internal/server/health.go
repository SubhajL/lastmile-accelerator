package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthDeps struct {
	DB          *sql.DB
	RedisClient *redis.Client
	NATSReady   func() error
	VaultHealth func(context.Context) error
	DBPing      func(context.Context, *sql.DB) error
	RedisPing   func(context.Context, *redis.Client) error
}

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func NewHealthHandler(deps *HealthDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		checks := make(map[string]string)
		healthy := true

		// Check database
		if deps.DBPing != nil {
			if err := deps.DBPing(ctx, deps.DB); err != nil {
				checks["database"] = err.Error()
				healthy = false
			} else {
				checks["database"] = "ok"
			}
		}

		// Check Redis
		if deps.RedisPing != nil {
			if err := deps.RedisPing(ctx, deps.RedisClient); err != nil {
				checks["redis"] = err.Error()
				healthy = false
			} else {
				checks["redis"] = "ok"
			}
		}

		// Check NATS
		if deps.NATSReady != nil {
			if err := deps.NATSReady(); err != nil {
				checks["nats"] = err.Error()
				healthy = false
			} else { checks["nats"] = "ok" }
		}

		// Check Vault
		if deps.VaultHealth != nil {
			if err := deps.VaultHealth(ctx); err != nil {
				checks["vault"] = err.Error()
				healthy = false
			} else { checks["vault"] = "ok" }
		}

		response := HealthResponse{
			Checks: checks,
		}

		w.Header().Set("Content-Type", "application/json")

		if healthy {
			response.Status = "ok"
			w.WriteHeader(http.StatusOK)
		} else {
			response.Status = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(response)
	}
}
