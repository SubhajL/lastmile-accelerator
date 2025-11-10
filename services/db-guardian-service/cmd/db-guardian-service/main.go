package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/cache"
	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/database"
	"example.com/lma/db-guardian-service/internal/events"
	"example.com/lma/db-guardian-service/internal/repository"
	"example.com/lma/db-guardian-service/internal/secrets"
	"example.com/lma/db-guardian-service/internal/server"
	"example.com/lma/db-guardian-service/internal/service"
	"example.com/lma/db-guardian-service/internal/telemetry"
	"example.com/lma/db-guardian-service/pkg/logger"
	"example.com/lma/db-guardian-service/internal/auth"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg.ServiceName, cfg.Environment, nil)
	log.Info("Starting db-guardian-service",
		logger.Field{Key: "environment", Value: cfg.Environment},
		logger.Field{Key: "port", Value: cfg.ServicePort},
	)

	// Initialize telemetry
	shutdownTracer, err := telemetry.InitTracer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}
	defer func() {
		if err := shutdownTracer(context.Background()); err != nil {
			log.Error("Failed to shutdown tracer", logger.Field{Key: "error", Value: err.Error()})
		}
	}()

	// Initialize database (optional)
	var db *sql.DB
	if cfg.DatabaseURL != "" {
		dbConn, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Error("Failed to connect to database", logger.Field{Key: "error", Value: err.Error()})
			// Not fatal - service can start without DB
		} else {
			defer dbConn.Close()
			db = dbConn
			log.Info("Database connection established")
		}
	}

	// Initialize Redis (optional)
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient, err = cache.NewRedisClient(cfg.RedisAddr)
		if err != nil {
			log.Error("Failed to connect to Redis", logger.Field{Key: "error", Value: err.Error()})
			// Not fatal - service can start without Redis
		} else {
			defer redisClient.Close()
			log.Info("Redis connection established")
		}
	}

	// Initialize NATS (optional)
	var natsConn *nats.Conn
	if cfg.NATSUrl != "" {
		natsConn, err = events.NewNATSClient(cfg.NATSUrl)
		if err != nil {
			log.Error("Failed to connect to NATS", logger.Field{Key: "error", Value: err.Error()})
			// Not fatal - service can start without NATS
		} else {
			defer natsConn.Close()
			log.Info("NATS connection established")
		}
	}

	// Initialize Vault (optional)
	if cfg.VaultAddr != "" && cfg.VaultRoleID != "" && cfg.VaultSecretID != "" {
		vaultCfg := &secrets.VaultConfig{
			Address:  cfg.VaultAddr,
			RoleID:   cfg.VaultRoleID,
			SecretID: cfg.VaultSecretID,
		}
		_, err = secrets.NewVaultClient(vaultCfg)
		if err != nil {
			log.Error("Failed to connect to Vault", logger.Field{Key: "error", Value: err.Error()})
			// Not fatal - service can start without Vault
		} else {
			log.Info("Vault connection established")
		}
	}

	// Wire analyzers and services (if DB available)
	var connSvc *service.ConnectionService
	var analysisSvc *service.AnalysisService
	var migrationGuard *analyzer.MigrationGuard
	var recsSvc *service.RecommendationsService
	var policySvc *service.PolicyService
	var driftSvc *service.DriftService

	// Wire analyzers and services (if DB available)
	if db != nil {
		inspector := analyzer.NewPostgresInspector(db)
		roleAnalyzer := analyzer.NewRoleAnalyzer(inspector)
		migrationGuard = analyzer.NewMigrationGuard(inspector)
		indexAdvisor := analyzer.NewIndexAdvisor(inspector)

		connSvc = service.NewConnectionService(db)
		analysisSvc = service.NewAnalysisService(db, natsConn, roleAnalyzer, migrationGuard, indexAdvisor)

		// Additional services for gRPC v1
		idxRepo := repository.NewIndexRecommendationsRepository(db)
		recsSvc = service.NewRecommendationsService(idxRepo, roleAnalyzer)
		policyRepo := repository.NewRolePoliciesRepository(db)
		policySvc = service.NewPolicyService(policyRepo)
		driftSvc = service.NewDriftService(inspector, idxRepo)

		log.Info("Analyzers and services wired")
	}

	// Create HTTP server
	deps := &server.Dependencies{
		DB:          db,
		RedisClient: redisClient,
		NATSConn:    natsConn,
		Logger:      log,
		ConnSvc:     connSvc,
		AnalysisSvc: analysisSvc,
		MigGuard:    migrationGuard,

		// gRPC v1 providers
		RecsProvider: recsSvc,
		PolicyMgr:    policySvc,
		DriftCheck:   driftSvc,
	}

	// Enable simple auth for gRPC (dev). Replace with real JWT/JWKS in prod.
	deps.Authenticator = auth.NewSimpleAuthenticator()
	srv := server.New(cfg, deps)

	// Start gRPC (if enabled via build tag)
	stopGRPC, err := maybeStartGRPC(ctx, deps, log)
	if err != nil {
		log.Error("Failed to start gRPC server", logger.Field{Key: "error", Value: err.Error()})
	} else {
		defer stopGRPC()
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down", logger.Field{Key: "signal", Value: sig.String()})
		cancel()
		// Wait for server to shut down
		<-errChan
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	log.Info("Service stopped gracefully")
	return nil
}
