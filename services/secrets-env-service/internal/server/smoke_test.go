package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/handlers"
	appLogger "example.com/lma/secrets-env-service/internal/logger"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/service"
	"example.com/lma/secrets-env-service/internal/vault"
)

type smokeFakeVerifier struct{ allow bool; claims handlers.Claims; err error }

func (f *smokeFakeVerifier) Verify(_ context.Context, _ string) (*handlers.Claims, error) {
	if !f.allow { return nil, f.err }
	return &f.claims, nil
}

type smokeCapturePublisher struct{ topics []string }

func (c *smokeCapturePublisher) Publish(ctx context.Context, topic string, payload any) error {
	c.topics = append(c.topics, topic)
	return nil
}

type smokeSecretsAdapter struct{ s *service.SecretsService }

func (a smokeSecretsAdapter) CreateSecret(ctx any, secret *domain.Secret, value map[string]interface{}) error {
	return a.s.CreateSecret(ctx.(context.Context), secret, value)
}
func (a smokeSecretsAdapter) GetSecret(ctx any, tenantID, projectID, key, environment string) (*domain.Secret, map[string]interface{}, error) {
	return a.s.GetSecret(ctx.(context.Context), tenantID, projectID, key, environment)
}
func (a smokeSecretsAdapter) DeleteSecret(ctx any, tenantID, projectID, key, environment string) error {
	return a.s.DeleteSecret(ctx.(context.Context), tenantID, projectID, key, environment)
}
func (a smokeSecretsAdapter) ListSecrets(ctx any, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
	return a.s.ListSecrets(ctx.(context.Context), projectID, environment, limit, cursor)
}

func buildSecretsStack(t *testing.T, ver *smokeFakeVerifier) (http.Handler, *smokeCapturePublisher) {
	t.Helper()
	log := appLogger.New("test", "info", new(bytes.Buffer))
	secretsRepo := repository.NewSecretsRepository(nil)
	v := &vault.Client{}
	v.SetTestMode(true)
	cap := &smokeCapturePublisher{}
svc := service.NewSecretsService(v, secretsRepo, nil, cap)
	h := handlers.NewSecretsHandler(smokeSecretsAdapter{s: svc})
	r := SetupRoutes(h, nil, nil, handlers.PanicRecovery(log), handlers.RequestLogger(log), handlers.JWTAuth(ver))
	return r, cap
}

func TestSmoke_Healthz_OK(t *testing.T) {
	log := appLogger.New("test", "info", new(bytes.Buffer))
	r := SetupRoutes(nil, nil, nil, handlers.PanicRecovery(log), handlers.RequestLogger(log))
	ts := httptest.NewServer(r)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil { t.Fatalf("healthz req err: %v", err) }
	if resp.StatusCode != http.StatusOK { t.Fatalf("want 200, got %d", resp.StatusCode) }
}

func TestSmoke_SecretsCreate_FullStack_401_403_201(t *testing.T) {
	// 401 without Authorization header
	r, _ := buildSecretsStack(t, &smokeFakeVerifier{allow: false})
	ts := httptest.NewServer(r)
	defer ts.Close()
	b, _ := json.Marshal(map[string]any{
		"key": "API_KEY", "environment": "prod", "value": map[string]any{"k": "v"}, "createdBy": "user@example.com",
	})
	resp, _ := http.Post(ts.URL+"/v1/projects/p/secrets", "application/json", bytes.NewReader(b))
	if resp.StatusCode != http.StatusUnauthorized { t.Fatalf("want 401, got %d", resp.StatusCode) }

	// 403 with missing scope
	r2, _ := buildSecretsStack(t, &smokeFakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:read"}}})
	ts2 := httptest.NewServer(r2)
	defer ts2.Close()
	req, _ := http.NewRequest(http.MethodPost, ts2.URL+"/v1/projects/p/secrets", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ok")
	req.Header.Set("X-Tenant-ID", "t1")
	resp2, _ := http.DefaultClient.Do(req)
	if resp2.StatusCode != http.StatusForbidden { t.Fatalf("want 403, got %d", resp2.StatusCode) }

	// 201 with correct scope and tenant, and publish
	verYes := &smokeFakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:write"}}}
	r3, cap := buildSecretsStack(t, verYes)
	ts3 := httptest.NewServer(r3)
	defer ts3.Close()
	req3, _ := http.NewRequest(http.MethodPost, ts3.URL+"/v1/projects/p/secrets", bytes.NewReader(b))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer ok")
	req3.Header.Set("X-Tenant-ID", "t1")
	resp3, _ := http.DefaultClient.Do(req3)
	if resp3.StatusCode != http.StatusCreated { t.Fatalf("want 201, got %d", resp3.StatusCode) }
	if len(cap.topics) == 0 || cap.topics[0] != "secret.created" { t.Fatalf("expected secret.created publish, got %v", cap.topics) }
}
