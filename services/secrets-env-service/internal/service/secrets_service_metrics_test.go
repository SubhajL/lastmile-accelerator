package service

import (
    "context"
    "testing"

    "example.com/lma/secrets-env-service/internal/domain"
    "example.com/lma/secrets-env-service/internal/metrics"
    "example.com/lma/secrets-env-service/internal/repository"
    "example.com/lma/secrets-env-service/internal/vault"
    "github.com/prometheus/client_golang/prometheus"
    dto "github.com/prometheus/client_model/go"
)

func fam(reg *prometheus.Registry, name string) *dto.MetricFamily {
    fams, err := reg.Gather()
    if err != nil { return nil }
    for _, f := range fams { if f.GetName() == name { return f } }
    return nil
}

func TestSecretsMetrics_CreateSuccess(t *testing.T) {
    reg := prometheus.NewRegistry()
    _ = metrics.Init(reg)
    v := &vault.Client{}; v.SetTestMode(true)
    repo := repository.NewSecretsRepository(nil)
    svc := NewSecretsService(v, repo, nil, nil)
    sec := &domain.Secret{ID:"1", TenantID:"t1", ProjectID:"p1", Key:"K", Environment:"dev", CreatedBy:"u"}
    if err := svc.CreateSecret(context.Background(), sec, map[string]any{"x":"y"}); err != nil { t.Fatalf("create err: %v", err) }
    mf := fam(reg, "secrets_operations_total")
    if mf == nil { t.Fatalf("metric family not found") }
    found := false
    for _, m := range mf.GetMetric() {
        var op, outcome string
        for _, lp := range m.GetLabel() {
            switch lp.GetName() { case "op": op = lp.GetValue(); case "outcome": outcome = lp.GetValue() }
        }
        if op == "create" && outcome == "success" && m.GetCounter().GetValue() >= 1 { found = true; break }
    }
    if !found { t.Fatalf("expected secrets create success counter >=1") }
}

func TestSecretsMetrics_GetError(t *testing.T) {
    reg := prometheus.NewRegistry()
    _ = metrics.Init(reg)
    v := &vault.Client{}; v.SetTestMode(true)
    repo := repository.NewSecretsRepository(nil)
    svc := NewSecretsService(v, repo, nil, nil)
    _, _, _ = svc.GetSecret(context.Background(), "t1", "p1", "missing", "dev")
    mf := fam(reg, "secrets_operations_total")
    if mf == nil { t.Fatalf("metric family not found") }
    found := false
    for _, m := range mf.GetMetric() {
        var op, outcome string
        for _, lp := range m.GetLabel() {
            switch lp.GetName() { case "op": op = lp.GetValue(); case "outcome": outcome = lp.GetValue() }
        }
        if op == "get" && outcome == "error" && m.GetCounter().GetValue() >= 1 { found = true; break }
    }
    if !found { t.Fatalf("expected secrets get error counter >=1") }
}
