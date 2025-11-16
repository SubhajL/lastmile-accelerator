package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/auth"
	"example.com/lma/db-guardian-service/internal/ratelimit"
	"example.com/lma/db-guardian-service/pkg/logger"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	httpServer *http.Server
	handler    http.Handler
	config     *config.Config
	logger     *logger.Logger
}

// Local aliases to avoid circular imports in server package
// These are interfaces defined in external packages (auth, ratelimit)
// referenced by Dependencies for middleware wiring.

type Authenticator = auth.Authenticator

type RateLimiter = ratelimit.RateLimiter

type Dependencies struct {
	// ... existing fields above ...
	DB          *sql.DB
	RedisClient *redis.Client
	NATSConn    *nats.Conn
	Logger      *logger.Logger

	// Optional services for API endpoints
	ConnSvc     ConnectionManager
	AnalysisSvc AnalysisRunner
	MigGuard    MigrationValidator

// Middlewares
	Authenticator Authenticator
	RateLimiter   RateLimiter

	// v1 providers
	PolicyMgr    PolicyManager
	RecsProvider RecommendationsProvider
	DriftCheck   DriftChecker

	// Optional Vault health (nil ok)
	VaultHealth func(ctx context.Context) error
}

func New(cfg *config.Config, deps *Dependencies) *Server {
	mux := http.NewServeMux()

	// Register health check endpoint (expanded later)
    var dbRef *sql.DB
    var redisRef *redis.Client
    if deps != nil { dbRef = deps.DB; redisRef = deps.RedisClient }
    healthDeps := &HealthDeps{
        DB:          dbRef,
        RedisClient: redisRef,
        DBPing:      defaultDBPing,
        RedisPing:   defaultRedisPing,
    }
	mux.HandleFunc("/healthz", NewHealthHandler(healthDeps))

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

    // Register API endpoints if services provided
    registerAPI(mux, cfg, deps)
    registerAPIv1(mux, cfg, deps)

    // Global middlewares: auth scopes (if configured) and http metrics
    var authn Authenticator
    if deps != nil { authn = deps.Authenticator }
    base := Chain(mux, RequireScopesFunc(authn, HTTPScopeResolver), Metrics())
    // Wrap with OpenTelemetry instrumentation
    handler := otelhttp.NewHandler(base, "db-guardian-service")

	srv := &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.ServicePort,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		handler: handler,
		config:  cfg,
        logger:  func() *logger.Logger { if deps != nil { return deps.Logger }; return nil }(),
	}

	return srv
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if s.logger != nil {
			s.logger.Info(fmt.Sprintf("Starting server on port %s", s.config.ServicePort))
		}
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for context cancellation
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if s.logger != nil {
			s.logger.Info("Shutting down server gracefully")
		}

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		return nil
	}
}

// Default ping functions
func defaultDBPing(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return nil // Skip if not configured
	}
	return db.PingContext(ctx)
}

func defaultRedisPing(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return nil // Skip if not configured
	}
	return client.Ping(ctx).Err()
}
