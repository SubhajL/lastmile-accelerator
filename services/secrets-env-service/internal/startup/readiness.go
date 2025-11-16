package startup

import (
    "context"
    "database/sql"
    "net/url"
    "sync"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "example.com/lma/secrets-env-service/internal/config"
    "example.com/lma/secrets-env-service/internal/vault"
)

// Status is a threadsafe readiness state with optional details.
type Status struct {
    mu      sync.RWMutex
    ready   bool
    details map[string]any
}

func NewStatus() *Status { return &Status{details: map[string]any{}} }
func (s *Status) Ready() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.ready }
func (s *Status) Set(v bool) { s.mu.Lock(); s.ready = v; s.mu.Unlock() }
func (s *Status) SetDetail(k string, v any) { s.mu.Lock(); s.details[k] = v; s.mu.Unlock() }
func (s *Status) Snapshot() map[string]any { s.mu.RLock(); defer s.mu.RUnlock(); out := make(map[string]any, len(s.details)); for k,v := range s.details { out[k]=v }; out["ready"]=s.ready; return out }

var global = NewStatus()

func SetReady(v bool) { global.Set(v) }
func IsReady() bool { return global.Ready() }
func Details() map[string]any { return global.Snapshot() }

// Run executes startup checks; returns final status and any critical error.
func Run(ctx context.Context, cfg *config.Config, v *vault.Client) (*Status, error) {
    st := NewStatus()
    criticalTO := time.Duration(cfg.Startup.CriticalTimeoutS) * time.Second
    if criticalTO <= 0 { criticalTO = 3 * time.Second }
    optionalTO := time.Duration(cfg.Startup.OptionalTimeoutS) * time.Second
    if optionalTO <= 0 { optionalTO = 1 * time.Second }

    // Critical: Postgres (if enabled)
    if cfg.Database.UsePGRepos {
        cctx, cancel := context.WithTimeout(ctx, criticalTO)
        err := checkPostgres(cctx, cfg.Database.URL)
        cancel()
        if err != nil { st.Set(false); st.SetDetail("postgres", err.Error()); return st, err }
        st.SetDetail("postgres", "ok")
    }

    // Critical: Vault
    {
        cctx, cancel := context.WithTimeout(ctx, criticalTO)
        err := v.HealthCheck(cctx)
        cancel()
        if err != nil { st.Set(false); st.SetDetail("vault", err.Error()); return st, err }
        st.SetDetail("vault", "ok")
    }

    // Optional: Redis
    if cfg.Redis.URL != "" {
        cctx, cancel := context.WithTimeout(ctx, optionalTO)
        _ = cctx
        if _, err := url.ParseRequestURI(cfg.Redis.URL); err != nil {
            st.SetDetail("redis", "invalid url")
        } else {
            st.SetDetail("redis", "configured")
        }
        cancel()
    }
    // Optional: NATS
    if cfg.NATS.URL != "" {
        if _, err := url.ParseRequestURI(cfg.NATS.URL); err != nil {
            st.SetDetail("nats", "invalid url")
        } else {
            st.SetDetail("nats", "configured")
        }
    }
    // Optional: S3 (configuration validation only)
    if cfg.Storage.S3.Endpoint != "" || cfg.Storage.S3.Bucket != "" {
        st.SetDetail("s3", "configured")
    }

    st.Set(true)
    // Update global
    SetReady(true)
    return st, nil
}

func checkPostgres(ctx context.Context, dsn string) error {
    db, err := sql.Open("pgx", dsn)
    if err != nil { return err }
    defer func(){ _ = db.Close() }()
    return db.PingContext(ctx)
}
