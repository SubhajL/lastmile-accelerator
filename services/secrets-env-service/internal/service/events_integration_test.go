package service

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/vault"
)

type capturePublisher struct{ topics []string }

func (c *capturePublisher) Publish(ctx context.Context, topic string, payload any) error {
	c.topics = append(c.topics, topic)
	return nil
}

func TestSecretsService_PublishesOnCreate(t *testing.T) {
	v := &vault.Client{}
	v.SetTestMode(true)
	repo := repository.NewSecretsRepository(nil)
	cap := &capturePublisher{}
svc := NewSecretsService(v, repo, nil, cap)
	sec := &domain.Secret{ID: "1", TenantID: "t", ProjectID: "p", Key: "K", Environment: "dev", CreatedBy: "u", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = svc.CreateSecret(context.Background(), sec, map[string]any{"v":"x"})
	if len(cap.topics) == 0 || cap.topics[0] != "secret.created" {
		t.Fatalf("expected secret.created event, got %v", cap.topics)
	}
}
