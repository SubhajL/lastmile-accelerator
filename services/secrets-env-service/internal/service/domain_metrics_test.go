package service

import (
	"context"
	"errors"
	"testing"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/metrics"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/vault"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

type stubStorage struct {
	files []fileBlob
	err   error
}

func (s *stubStorage) ListFiles(ctx context.Context, projectID, snapshotID string) ([]fileBlob, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.files, nil
}

func TestDomainMetrics_SecretsParityLeaks(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics.Init(reg)

	t.Run("secrets create + get error", func(t *testing.T) {
		v := &vault.Client{}
		v.SetTestMode(true)
		repo := repository.NewSecretsRepository(nil)
		svc := NewSecretsService(v, repo, nil, nil)

		sec := &domain.Secret{ID: "1", TenantID: "t1", ProjectID: "p1", Key: "K", Environment: "dev", CreatedBy: "u"}
		require.NoError(t, svc.CreateSecret(context.Background(), sec, map[string]any{"x": "y"}))
		_, _, _ = svc.GetSecret(context.Background(), "t1", "p1", "missing", "dev")

		require.True(t, hasCounter(reg, "secrets_operations_total", map[string]string{"op": "create", "outcome": "success"}))
		require.True(t, hasCounter(reg, "secrets_operations_total", map[string]string{"op": "get", "outcome": "error"}))
	})

	t.Run("parity compute + latest + history", func(t *testing.T) {
		secretsRepo := repository.NewSecretsRepository(nil)
		parityRepo := repository.NewParityRepository()
		svc := NewParityService(parityRepo, secretsRepo, nil)

		_, err := svc.CheckParity(context.Background(), "p1", "dev", "prod")
		require.NoError(t, err)
		_, err = svc.GetLatestCheck(context.Background(), "p1")
		require.NoError(t, err)
		_, err = svc.GetCheckHistory(context.Background(), "p1", 10)
		require.NoError(t, err)

		require.True(t, hasCounter(reg, "parity_operations_total", map[string]string{"op": "compute", "outcome": "success"}))
		require.True(t, hasCounter(reg, "parity_operations_total", map[string]string{"op": "latest", "outcome": "success"}))
		require.True(t, hasCounter(reg, "parity_operations_total", map[string]string{"op": "history", "outcome": "success"}))
	})

	t.Run("leak scan success + error", func(t *testing.T) {
		repo := repository.NewLeakScanRepository()

		svc := NewLeakScanService(repo, &stubStorage{files: []fileBlob{{Path: "/a", Content: []byte("hello")}}}, nil)
		_, err := svc.ScanSnapshot(context.Background(), "p1", "s1")
		require.NoError(t, err)

		svc2 := NewLeakScanService(repo, &stubStorage{err: errors.New("boom")}, nil)
		_, _ = svc2.ScanSnapshot(context.Background(), "p1", "s2")

		require.True(t, hasCounter(reg, "leak_operations_total", map[string]string{"op": "scan", "outcome": "success"}))
		require.True(t, hasCounter(reg, "leak_operations_total", map[string]string{"op": "scan", "outcome": "error"}))
	})
}

func hasCounter(reg *prometheus.Registry, metricName string, wantLabels map[string]string) bool {
	fams, err := reg.Gather()
	if err != nil {
		return false
	}

	for _, f := range fams {
		if f.GetName() != metricName {
			continue
		}
		for _, m := range f.GetMetric() {
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}

			matches := true
			for k, v := range wantLabels {
				if labels[k] != v {
					matches = false
					break
				}
			}

			if matches && m.GetCounter().GetValue() >= 1 {
				return true
			}
		}
	}

	return false
}
