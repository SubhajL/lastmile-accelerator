package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	"example.com/lma/observability-service/internal/config"
	"example.com/lma/observability-service/internal/health"
	"example.com/lma/observability-service/internal/storage"
	"example.com/lma/observability-service/internal/telemetry"

	"example.com/lma/observability-service/internal/grpcserver"
	"example.com/lma/observability-service/internal/handlers"
	"example.com/lma/observability-service/internal/metrics"
	"example.com/lma/observability-service/internal/messaging"
	"example.com/lma/observability-service/internal/middleware"
	"example.com/lma/observability-service/internal/repository"
	"example.com/lma/observability-service/internal/scheduler"
	"example.com/lma/observability-service/internal/services"
	"example.com/lma/observability-service/internal/traces"
	"example.com/lma/observability-service/internal/logs"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func run() error {
	// Root context for background workers (scheduler)
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log.Printf("starting %s in %s environment", cfg.ServiceName, cfg.Env)

	// Initialize telemetry
	telProvider, err := telemetry.InitTelemetry(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telProvider.Shutdown(ctx); err != nil {
			log.Printf("error shutting down telemetry: %v", err)
		}
	}()

	log.Println("telemetry initialized")

	// Initialize database connection
	db, err := storage.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	log.Println("database connected")

	// Run DB migrations
	if err := storage.RunMigrations(context.Background(), db); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}
	log.Println("migrations applied")

	// Initialize Redis connection
	redisClient, err := storage.NewRedisClient(cfg.RedisURL, "", 0)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisClient.Close()
	log.Println("redis connected")

	// Initialize health checker
	healthChecker := health.NewChecker(db, redisClient)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthChecker.HTTPHandler())
	mux.HandleFunc("/readyz", healthChecker.HTTPHandler())

// Auth verifier
verifier := middleware.NewJWKSVerifier(cfg.JWKSURL, cfg.JWTIssuer, cfg.JWTAudience, "scopes", nil)

// Wire OTel endpoints
otelRepo := repository.NewOTelRepository(db)
otelSvc := services.NewOTelService(otelRepo, telProvider.Tracer())
otelHandler := handlers.NewOTelHandler(otelSvc)
// GET /v1/otel/presets (read)
mux.Handle("/v1/otel/presets", middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(otelHandler.GetPresets))))
// GET /v1/otel/presets/{framework} (read)
mux.Handle("/v1/otel/presets/", middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(otelHandler.GetPresetByFramework))))

	// Prepare SLO service and handler
sloRepo := repository.NewSLORepository(db)
prom := metrics.NewPromClient(cfg.PrometheusURL, &http.Client{Timeout: 5 * time.Second})
sloSvc := services.NewSLOService(sloRepo, telProvider.Tracer(), prom)
sloHandler := handlers.NewSLOHandler(sloSvc)

	// Install scheduler metrics recorder (OTel-backed)
	scheduler.UseMetricsRecorder(scheduler.NewOTelRecorder(telProvider.Meter()))

	// Start SLO evaluator scheduler with configurable env
slEval := scheduler.NewSLOEvaluator(sloRepo, sloSvc)
interval := 1 * time.Minute
if d, err := time.ParseDuration(strings.TrimSpace(cfg.SLOEvalInterval)); err == nil && d > 0 { interval = d }
workers := 4
if w, err := strconv.Atoi(strings.TrimSpace(cfg.SLOEvalWorkers)); err == nil && w > 0 { workers = w }
go func(){ _ = slEval.Run(rootCtx, interval, workers) }()

// Traces/Logs queries
trCli := traces.NewClient(cfg.TempoURL, &http.Client{Timeout: 10 * time.Second})
logCli := logs.NewClient(cfg.LokiURL, &http.Client{Timeout: 10 * time.Second})
queriesSvc := services.NewObservabilityQueries(trCli, logCli, prom)
queriesHandler := handlers.NewQueriesHandler(queriesSvc)

mux.HandleFunc("/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		// Routes: /v1/projects/{id}/otel/config (GET, POST)
		if strings.HasSuffix(r.URL.Path, "/otel/config") {
			if r.Method == http.MethodPost {
				middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(otelHandler.ApplyConfigToProject))).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodGet {
				middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(otelHandler.GetProjectConfig))).ServeHTTP(w, r)
				return
			}
		}
		// SLO routes under /v1/projects/{id}/slos
		if strings.Contains(r.URL.Path, "/slos") && len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) >= 4 {
			if r.Method == http.MethodPost {
				middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(sloHandler.CreateSLO))).ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodGet {
				middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(sloHandler.ListProjectSLOs))).ServeHTTP(w, r)
				return
			}
		}
		// Traces search
		if strings.HasSuffix(r.URL.Path, "/traces") && r.Method == http.MethodGet {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(queriesHandler.SearchTraces))).ServeHTTP(w, r); return
		}
		// Logs search
		if strings.HasSuffix(r.URL.Path, "/logs") && r.Method == http.MethodGet {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(queriesHandler.SearchLogs))).ServeHTTP(w, r); return
		}
		// Golden dashboard
		if strings.HasSuffix(r.URL.Path, "/dashboards/golden") && r.Method == http.MethodGet {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(queriesHandler.GoldenDashboard))).ServeHTTP(w, r); return
		}
		w.WriteHeader(http.StatusNotFound)
	})

// SLO Endpoints outside projects scope
mux.HandleFunc("/v1/slos/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(sloHandler.GetSLOStatus))).ServeHTTP(w, r); return
		}
		if strings.HasSuffix(r.URL.Path, "/history") {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(sloHandler.GetSLOHistory))).ServeHTTP(w, r); return
		}
		switch r.Method {
		case http.MethodGet:
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(sloHandler.GetSLO))).ServeHTTP(w, r)
		case http.MethodPut:
			middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(sloHandler.UpdateSLO))).ServeHTTP(w, r)
		case http.MethodDelete:
			middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(sloHandler.DeleteSLO))).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

// Error inbox endpoints
errRepo := repository.NewErrorRepository(db)
errSvc := services.NewErrorService(errRepo, telProvider.Tracer())
errHandler := handlers.NewErrorHandler(errSvc)
mux.HandleFunc("/v1/projects/", func(w http.ResponseWriter, r *http.Request){
		// append to existing /v1/projects handler: error inbox
		if strings.HasSuffix(r.URL.Path, "/errors/ingest") && r.Method == http.MethodPost {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:ingest")(http.HandlerFunc(errHandler.Ingest))).ServeHTTP(w, r); return
		}
		if strings.HasSuffix(r.URL.Path, "/errors") && r.Method == http.MethodGet {
			middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(errHandler.ListGroups))).ServeHTTP(w, r); return
		}
		w.WriteHeader(http.StatusNotFound)
	})
// Error group routes
mux.Handle("/v1/errors/groups/", middleware.JWT(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
	if strings.HasSuffix(r.URL.Path, "/resolve") && r.Method==http.MethodPost {
		middleware.RequireScopes("observability:write")(http.HandlerFunc(errHandler.ResolveGroup)).ServeHTTP(w, r); return
	}
	if strings.HasSuffix(r.URL.Path, "/events") && r.Method==http.MethodGet {
		middleware.RequireScopes("observability:read")(http.HandlerFunc(errHandler.ListGroupEvents)).ServeHTTP(w, r); return
	}
	if r.Method==http.MethodGet {
		middleware.RequireScopes("observability:read")(http.HandlerFunc(errHandler.GetGroup)).ServeHTTP(w, r); return
	}
	w.WriteHeader(http.StatusNotFound)
})))

// Alert endpoints
alertRepo := repository.NewAlertRepository(db)
nc, err := nats.Connect(cfg.NATSURL)
if err != nil { return fmt.Errorf("nats connect: %w", err) }
defer nc.Close()
pub := messaging.NewNATSPublisher(nc)
alertSvc := services.NewAlertService(alertRepo, telProvider.Tracer(), pub)
alertHandler := handlers.NewAlertHandler(alertSvc)
mux.HandleFunc("/v1/slos/", func(w http.ResponseWriter, r *http.Request){
		if strings.Contains(r.URL.Path, "/alerts") {
			if r.Method == http.MethodPost { middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(alertHandler.CreateAlert))).ServeHTTP(w, r); return }
			if r.Method == http.MethodGet { middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(alertHandler.GetSLOAlerts))).ServeHTTP(w, r); return }
		}
		w.WriteHeader(http.StatusNotFound)
	})
mux.HandleFunc("/v1/alerts/", func(w http.ResponseWriter, r *http.Request){
		if strings.HasSuffix(r.URL.Path, "/history") { middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(alertHandler.GetAlertHistory))).ServeHTTP(w, r); return }
		switch r.Method {
		case http.MethodGet: middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(alertHandler.GetAlert))).ServeHTTP(w, r)
		case http.MethodPut: middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(alertHandler.UpdateAlert))).ServeHTTP(w, r)
		case http.MethodDelete: middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(alertHandler.DeleteAlert))).ServeHTTP(w, r)
		default: w.WriteHeader(http.StatusNotFound)
		}
	})

// Trace get
mux.Handle("/v1/traces/", middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(queriesHandler.GetTrace))))

srv := &http.Server{
		Addr:         ":" + cfg.ServicePort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

// Start gRPC server using JSON codec and auth interceptor
grpcserver.RegisterJSONCodec()
grps := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.UnaryAuthInterceptor(verifier)))
obsSrv := grpcserver.NewObservabilityServer(sloSvc, queriesSvc)
grpcserver.RegisterObservabilityServer(grps, obsSrv)

l, err := net.Listen("tcp", ":"+cfg.GRPCPort)
if err != nil { return fmt.Errorf("grpc listen: %w", err) }

go func(){
	log.Printf("%s gRPC listening on :%s", cfg.ServiceName, cfg.GRPCPort)
	if err := grps.Serve(l); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		log.Printf("grpc serve error: %v", err)
	}
}()

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("%s listening on :%s", cfg.ServiceName, cfg.ServicePort)
		serverErrors <- srv.ListenAndServe()
	}()

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until shutdown signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
case sig := <-shutdown:
		log.Printf("received signal: %v, starting graceful shutdown", sig)

		// Stop background workers
		rootCancel()

		// Give outstanding requests 30s to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
		// Force close if graceful shutdown fails
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
		// Stop gRPC server
		grps.GracefulStop()
}

	log.Println("server stopped gracefully")
	return nil
}
