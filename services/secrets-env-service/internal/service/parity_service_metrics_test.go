package service

import (
    "context"
    "testing"
    "time"

    "example.com/lma/secrets-env-service/internal/metrics"
    "example.com/lma/secrets-env-service/internal/repository"
    "github.com/prometheus/client_golang/prometheus"
)

func TestParityMetrics_ComputeAndLatest_Success(t *testing.T) {
    reg := prometheus.NewRegistry()
    _ = metrics.Init(reg)
    secretsRepo := repository.NewSecretsRepository(nil)
    parityRepo := repository.NewParityRepository()
    svc := NewParityService(parityRepo, secretsRepo, nil)

    // compute
    if _, err := svc.CheckParity(context.Background(), "p1", "dev", "prod"); err != nil {
        t.Fatalf("compute err: %v", err)
    }
    // latest (seeded by compute)
    if _, err := svc.GetLatestCheck(context.Background(), "p1"); err != nil {
        t.Fatalf("latest err: %v", err)
    }
    // history
    if _, err := svc.GetCheckHistory(context.Background(), "p1", 10); err != nil {
        t.Fatalf("history err: %v", err)
    }

    fam := fam(reg, "parity_operations_total")
    if fam == nil { t.Fatalf("parity family not found") }
    want := map[[2]string]bool{
        {"compute", "success"}: false,
        {"latest", "success"}:  false,
        {"history", "success"}: false,
    }
    for _, m := range fam.GetMetric() {
        var op, outcome string
        for _, lp := range m.GetLabel() {
            if lp.GetName() == "op" { op = lp.GetValue() }
            if lp.GetName() == "outcome" { outcome = lp.GetValue() }
        }
        key := [2]string{op, outcome}
        if _, ok := want[key]; ok && m.GetCounter().GetValue() >= 1 { want[key] = true }
    }
    // ensure all expected present
    for k, ok := range want { if !ok { t.Fatalf("missing metric %v", k) } }

    _ = time.Now() // silence import if unused in future changes
}
