package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"example.com/lma/secrets-env-service/internal/cache"
	"example.com/lma/secrets-env-service/internal/config"
	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/events"
	grpcapi "example.com/lma/secrets-env-service/internal/grpc"
	"example.com/lma/secrets-env-service/internal/handlers"
	"example.com/lma/secrets-env-service/internal/logger"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/security"
	"example.com/lma/secrets-env-service/internal/server"
	"example.com/lma/secrets-env-service/internal/service"
	"example.com/lma/secrets-env-service/internal/vault"
)

func main() {
	cfg, err := config.Load()
	if err != nil { panic(err) }

	log := logger.New(cfg.ServiceName, cfg.LogLevel, os.Stdout)

    // OpenTelemetry init (required when endpoint is configured)
    ctx := context.Background()
    shutdownOTel, err := initOTel(ctx, cfg.Observability, cfg.ServiceName)
    if err != nil { panic(err) }
    defer func() { _ = shutdownOTel(context.Background()) }()

	// Repositories
	var secretsRepo repository.SecretsMetaRepo
	var auditRepo *repository.AuditLogRepositoryPG
	var (
		parityRepoAny any
		leakRepoAny   any
	)
	if cfg.Database.UsePGRepos {
		db, err := sql.Open("pgx", cfg.Database.URL)
		if err != nil { panic(err) }
		db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeS) * time.Second)
		secretsRepo = repository.NewSecretsRepository(db)
		auditRepo = repository.NewAuditLogRepositoryPostgres(db)
		parityRepoAny = repository.NewParityRepositoryPostgres(db)
		leakRepoAny = repository.NewLeakScanRepositoryPostgres(db)
	} else {
		secretsRepo = repository.NewSecretsRepository(nil)
		parityRepoAny = repository.NewParityRepository()
		leakRepoAny = repository.NewLeakScanRepository()
	}
	// Optionally wrap with Redis cache if configured
	if cfg.Redis.URL != "" {
		if rdb, err := cache.NewRedisClient(cfg.Redis.URL); err == nil {
			secretsRepo = cache.NewSecretsCacheRepository(secretsRepo, rdb, 5*time.Minute)
		}
	}
	parityRepo := parityRepoAny.(*repository.ParityRepository)
	if prpg, ok := parityRepoAny.(*repository.ParityRepositoryPG); ok {
		_ = prpg // keep for type inference
	}
	leakRepo := leakRepoAny.(*repository.LeakScanRepository)
	if lrpg, ok := leakRepoAny.(*repository.LeakScanRepositoryPG); ok {
		_ = lrpg
	}

	// Vault client (test mode for dev wiring; replace with AppRole auth later)
	v := &vault.Client{}
	v.SetTestMode(true)

	// Eventing (optional)
	var publisher events.Publisher
	if cfg.NATS.URL != "" {
		if p, err := events.NewNATSPublisherFromURL(cfg.NATS.URL); err == nil {
			publisher = p
		}
	}

	secretsSvc := service.NewSecretsService(v, secretsRepo, auditRepo, publisher)
	paritySvc := service.NewParityService(parityRepo, secretsRepo, publisher)

	// Storage: S3 (optional wiring)
	var blob *service.S3Blob
	if cfg.Storage.S3.Endpoint != "" && cfg.Storage.S3.Bucket != "" {
		minioCli, err := service.NewMinioS3Client(cfg.Storage.S3.Endpoint, cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey, cfg.Storage.S3.UseTLS)
		if err == nil {
			blob = service.NewS3BlobWithClient(service.S3Config{
				Endpoint:       cfg.Storage.S3.Endpoint,
				Bucket:         cfg.Storage.S3.Bucket,
				Prefix:         cfg.Storage.S3.Prefix,
				UseTLS:         cfg.Storage.S3.UseTLS,
				IgnoreGlobs:    cfg.Storage.S3.IgnoreGlobs,
				SizeLimitBytes: cfg.Storage.S3.SizeLimitBytes,
			}, minioCli)
		}
	}
	leakSvc := service.NewLeakScanService(leakRepo, blob, publisher)

	secretsH := handlers.NewSecretsHandler(secretsServiceAdapter{s: secretsSvc})
	parityH := handlers.NewParityHandler(parityServiceAdapter{s: paritySvc})
	leakH := handlers.NewLeakScanHandler(leakScanServiceAdapter{s: leakSvc})

	// Middleware chain: PanicRecovery -> RequestLogger -> JWTAuth
	var verifier handlers.TokenVerifier
	if cfg.Auth.JWKSURL != "" {
		verifier = security.NewJWKSVerifier(cfg.Auth.JWKSURL, cfg.Auth.Audience, cfg.Auth.Issuer, time.Duration(cfg.Auth.JWKSCacheTTL)*time.Second)
	} else {
		verifier = &denyAllVerifier{}
	}
	mw := []func(http.Handler) http.Handler{
		handlers.PanicRecovery(log),
		handlers.TraceContext(),
		handlers.OtelTracing(),
		handlers.RequestLogger(log),
		handlers.JWTAuth(verifier),
		handlers.RBACAugment(),
	}
	// Add rate limiting middleware
	limiter := security.NewRateLimiter(float64(cfg.RateLimit.RequestsPerSec), cfg.RateLimit.Burst)
	mw = append(mw, handlers.RateLimitHTTP(limiter.Allow))

	r := server.SetupRoutes(secretsH, parityH, leakH, mw...)
	h := server.NewServer(":"+cfg.ServicePort, r)

	// TLS (mTLS) optional
	tlsCfg, err := security.LoadServerTLS(cfg.TLS)
	if err != nil { log.Error().Err(err).Msg("failed to load TLS config") }

	// Context for shutdown
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	// Start HTTP and gRPC servers
	go func() {
		if tlsCfg != nil {
			ln, err := net.Listen("tcp", h.Addr)
			if err != nil { _ = err; return }
			tlsLn := tls.NewListener(ln, tlsCfg)
			_ = h.Serve(tlsLn)
			return
		}
		_ = h.ListenAndServe()
	}()
	go func() {
		_ = grpcapi.StartGRPCServer(appCtx, ":50064", secretsSvc, paritySvc, verifier, log, tlsCfg)
	}()

    <-appCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = h.Shutdown(shutdownCtx)
}

type denyAllVerifier struct{}

func (d *denyAllVerifier) Verify(ctx context.Context, token string) (*handlers.Claims, error) { return nil, context.Canceled }

type secretsServiceAdapter struct{ s *service.SecretsService }

func (a secretsServiceAdapter) CreateSecret(ctx any, secret *domain.Secret, value map[string]interface{}) error {
	return a.s.CreateSecret(ctx.(context.Context), secret, value)
}
func (a secretsServiceAdapter) GetSecret(ctx any, tenantID, projectID, key, environment string) (*domain.Secret, map[string]interface{}, error) {
	return a.s.GetSecret(ctx.(context.Context), tenantID, projectID, key, environment)
}
func (a secretsServiceAdapter) DeleteSecret(ctx any, tenantID, projectID, key, environment string) error {
	return a.s.DeleteSecret(ctx.(context.Context), tenantID, projectID, key, environment)
}
func (a secretsServiceAdapter) ListSecrets(ctx any, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
	return a.s.ListSecrets(ctx.(context.Context), projectID, environment, limit, cursor)
}

type parityServiceAdapter struct{ s *service.ParityService }

func (a parityServiceAdapter) CheckParity(ctx any, projectID, baseEnv, compareEnv string) (*domain.EnvParityCheck, error) {
	return a.s.CheckParity(ctx.(context.Context), projectID, baseEnv, compareEnv)
}
func (a parityServiceAdapter) GetLatestCheck(ctx any, projectID string) (*domain.EnvParityCheck, error) {
	return a.s.GetLatestCheck(ctx.(context.Context), projectID)
}
func (a parityServiceAdapter) GetCheckHistory(ctx any, projectID string, limit int) ([]*domain.EnvParityCheck, error) {
	return a.s.GetCheckHistory(ctx.(context.Context), projectID, limit)
}

type leakScanServiceAdapter struct{ s *service.LeakScanService }

func (a leakScanServiceAdapter) ScanSnapshot(ctx any, projectID, snapshotID string) ([]*domain.ClientLeakScan, error) {
	return a.s.ScanSnapshot(ctx.(context.Context), projectID, snapshotID)
}
func (a leakScanServiceAdapter) GetScanResults(ctx any, projectID, snapshotID, severity string) ([]*domain.ClientLeakScan, error) {
	return a.s.GetScanResults(ctx.(context.Context), projectID, snapshotID, severity)
}
func (a leakScanServiceAdapter) MarkAsFixed(ctx any, scanID string) error {
	return a.s.MarkAsFixed(ctx.(context.Context), scanID)
}

