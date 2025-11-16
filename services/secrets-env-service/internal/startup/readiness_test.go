package startup

import (
    "context"
    "testing"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
    "example.com/lma/secrets-env-service/internal/config"
    "example.com/lma/secrets-env-service/internal/vault"
)

func TestStatus_DefaultNotReady_ThenReady(t *testing.T) {
    st := NewStatus()
    if st.Ready() { t.Fatalf("expected not ready by default") }
    st.Set(true)
    if !st.Ready() { t.Fatalf("expected ready after Set(true)") }
}

func TestRun_PGFailFast_ReturnsError(t *testing.T) {
    cfg := &config.Config{}
    cfg.Database.URL = "postgres://localhost:1/bad" // invalid port ensures fail fast
    cfg.Database.UsePGRepos = true
    cfg.Startup = config.StartupConfig{CriticalTimeoutS: 1, OptionalTimeoutS: 1}
    v := &vault.Client{}
    v.SetTestMode(true)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    _, err := Run(ctx, cfg, v)
    if err == nil { t.Fatalf("expected error when PG enabled and unreachable") }
}

func TestRun_VaultHealthOK_Succeeds(t *testing.T) {
    cfg := &config.Config{}
    cfg.Database.UsePGRepos = false
    cfg.Startup = config.StartupConfig{CriticalTimeoutS: 1, OptionalTimeoutS: 1}
    v := &vault.Client{}; v.SetTestMode(true)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    st, err := Run(ctx, cfg, v)
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if !st.Ready() { t.Fatalf("expected ready status true") }
}

func TestRun_OptionalDepsErrors_DoNotFail(t *testing.T) {
    cfg := &config.Config{}
    cfg.Database.UsePGRepos = false
    cfg.Redis.URL = "://bad" // invalid URL
    cfg.NATS.URL = ":://bad" // invalid URL
    cfg.Storage.S3.Endpoint = "bad-endpoint" // invalid, but optional
    cfg.Startup = config.StartupConfig{CriticalTimeoutS: 1, OptionalTimeoutS: 1}
    v := &vault.Client{}; v.SetTestMode(true)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    st, err := Run(ctx, cfg, v)
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if !st.Ready() { t.Fatalf("expected ready despite optional failures") }
}
