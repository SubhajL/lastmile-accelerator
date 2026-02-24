package server

import (
	"net/http"

	"example.com/lma/secrets-env-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type secretsHTTP interface {
	CreateSecret(http.ResponseWriter, *http.Request)
	ListSecrets(http.ResponseWriter, *http.Request)
	GetSecret(http.ResponseWriter, *http.Request)
	DeleteSecret(http.ResponseWriter, *http.Request)
}

type parityHTTP interface {
	CheckParity(http.ResponseWriter, *http.Request)
	GetLatestCheck(http.ResponseWriter, *http.Request)
	GetCheckHistory(http.ResponseWriter, *http.Request)
}

type leakHTTP interface {
	ScanSnapshot(http.ResponseWriter, *http.Request)
	GetScanResults(http.ResponseWriter, *http.Request)
	MarkAsFixed(http.ResponseWriter, *http.Request)
}

// SetupRoutes configures chi router with all endpoints
func SetupRoutes(secretsH secretsHTTP, parityH parityHTTP, leakH leakHTTP, middlewares ...func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	for _, mw := range middlewares { r.Use(mw) }
	r.Use(handlers.RequireJSONContentType())

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { handlers.HealthCheck(w, r) })
	// metrics (Prometheus)
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// v1 routes (handlers may be nil during partial wiring)
	r.Route("/v1", func(r chi.Router) {
		r.Route("/projects/{projectID}", func(r chi.Router) {
			// secrets (protected)
			r.With(handlers.RequireScopes("secrets:write"), handlers.TenantIsolation()).Post("/secrets", func(w http.ResponseWriter, r *http.Request) {
				if secretsH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				secretsH.CreateSecret(w, r)
			})
			r.With(handlers.RequireScopes("secrets:read"), handlers.TenantIsolation()).Get("/secrets", func(w http.ResponseWriter, r *http.Request) {
				if secretsH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				secretsH.ListSecrets(w, r)
			})
	r.With(handlers.RequireScopes("secrets:read"), handlers.TenantIsolation(), handlers.ValidateKeyParam()).Get("/secrets/{key}", func(w http.ResponseWriter, r *http.Request) {
				if secretsH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				secretsH.GetSecret(w, r)
			})
			r.With(handlers.RequireScopes("secrets:write"), handlers.TenantIsolation(), handlers.ValidateKeyParam()).Delete("/secrets/{key}", func(w http.ResponseWriter, r *http.Request) {
				if secretsH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				secretsH.DeleteSecret(w, r)
			})

			// parity (protected)
			r.With(handlers.RequireScopes("parity:compute"), handlers.TenantIsolation()).Post("/env-parity", func(w http.ResponseWriter, r *http.Request) {
				if parityH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				parityH.CheckParity(w, r)
			})
			r.With(handlers.RequireScopes("parity:read"), handlers.TenantIsolation()).Get("/env-parity/latest", func(w http.ResponseWriter, r *http.Request) {
				if parityH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				parityH.GetLatestCheck(w, r)
			})
			r.With(handlers.RequireScopes("parity:read"), handlers.TenantIsolation()).Get("/env-parity/history", func(w http.ResponseWriter, r *http.Request) {
				if parityH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				parityH.GetCheckHistory(w, r)
			})

			// client leak scan (protected)
			r.With(handlers.RequireScopes("leaks:scan"), handlers.TenantIsolation()).Post("/scan/client-leaks", func(w http.ResponseWriter, r *http.Request) {
				if leakH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				leakH.ScanSnapshot(w, r)
			})
			r.With(handlers.RequireScopes("leaks:read"), handlers.TenantIsolation()).Get("/scan/client-leaks/{snapshotID}", func(w http.ResponseWriter, r *http.Request) {
				if leakH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				leakH.GetScanResults(w, r)
			})
			r.With(handlers.RequireScopes("leaks:write"), handlers.TenantIsolation()).Patch("/scan/client-leaks/{scanID}/fix", func(w http.ResponseWriter, r *http.Request) {
				if leakH == nil { handlers.Error(w, http.StatusNotImplemented, "not wired", nil); return }
				leakH.MarkAsFixed(w, r)
			})
		})
	})

	return r
}
