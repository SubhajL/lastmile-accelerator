package service

import (
    "context"
    "testing"

    "example.com/lma/secrets-env-service/internal/metrics"
    "example.com/lma/secrets-env-service/internal/repository"
    "github.com/prometheus/client_golang/prometheus"
)

type stubStorage struct{ files []fileBlob; err error }

func (s *stubStorage) ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error) {
    if s.err != nil { return nil, s.err }
    return s.files, nil
}

func TestLeakMetrics_Scan_SuccessAndError(t *testing.T) {
    reg := prometheus.NewRegistry()
    _ = metrics.Init(reg)
    repo := repository.NewLeakScanRepository()
    // success path
    svc := NewLeakScanService(repo, &stubStorage{files: []fileBlob{{Path:"/a", Content: []byte("AKIA1234567890ABCDEF")}}}, nil)
    if _, err := svc.ScanSnapshot(context.Background(), "p1", "s1"); err != nil { t.Fatalf("scan err: %v", err) }
    // error path (storage failure)
    svc2 := NewLeakScanService(repo, &stubStorage{err: assertAnError()}, nil)
    _, _ = svc2.ScanSnapshot(context.Background(), "p1", "s2")

    fam := fam(reg, "leak_operations_total")
    if fam == nil { t.Fatalf("leak family not found") }
    want := map[[2]string]bool{
        {"scan", "success"}: false,
        {"scan", "error"}:   false,
    }
    for _, m := range fam.GetMetric() {
        var op, outcome string
        for _, lp := range m.GetLabel() {
            if lp.GetName() == "op" { op = lp.GetValue() }
            if lp.GetName() == "outcome" { outcome = lp.GetValue() }
        }
        k := [2]string{op, outcome}
        if _, ok := want[k]; ok && m.GetCounter().GetValue() >= 1 { want[k] = true }
    }
    for k, ok := range want { if !ok { t.Fatalf("missing metric %v", k) } }
}

// assertAnError returns a non-nil error without external deps.
func assertAnError() error { return errSentinel }

var errSentinel = &simpleError{"boom"}

type simpleError struct{ s string }

func (e *simpleError) Error() string { return e.s }
