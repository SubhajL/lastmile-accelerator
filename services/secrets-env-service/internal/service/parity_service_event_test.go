package service

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/repository"
)

type capPub struct{ topics []string }
func (c *capPub) Publish(ctx context.Context, topic string, payload any) error { c.topics = append(c.topics, topic); return nil }

func TestParityService_PublishDriftDetected(t *testing.T) {
	secretsRepo := repository.NewSecretsRepository(nil)
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"1", ProjectID:"p", Key:"A", Environment:"dev", CreatedAt: time.Now()})
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"2", ProjectID:"p", Key:"B", Environment:"prod", CreatedAt: time.Now()})
	cap := &capPub{}
	svc := NewParityService(repository.NewParityRepository(), secretsRepo, cap)
	_, _ = svc.CheckParity(context.Background(), "p", "dev", "prod")
	found := false
	for _, tpc := range cap.topics { if tpc == "env.parity.drift.detected" { found = true; break } }
	if !found { t.Fatalf("expected env.parity.drift.detected publish, got %v", cap.topics) }
}
