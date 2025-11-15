package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "example.com/lma/secrets-env-service/internal/domain"
    "github.com/go-chi/chi/v5"
)

type secretsSvcStub struct{}

func (s *secretsSvcStub) CreateSecret(ctx any, secret *domain.Secret, value map[string]interface{}) error { return nil }
func (s *secretsSvcStub) GetSecret(ctx any, tenantID, projectID, key, environment string) (*domain.Secret, map[string]interface{}, error) {
    return &domain.Secret{}, map[string]interface{}{}, nil
}
func (s *secretsSvcStub) DeleteSecret(ctx any, tenantID, projectID, key, environment string) error { return nil }
func (s *secretsSvcStub) ListSecrets(ctx any, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
    return []*domain.Secret{}, "", nil
}

func TestCreateSecret_RejectsDisallowedEnvironment(t *testing.T) {
    SetAllowedEnvs([]string{"dev","prod"})
    defer SetAllowedEnvs([]string{"dev","staging","prod","production"})
    h := NewSecretsHandler(&secretsSvcStub{})
    r := chi.NewRouter()
    r.Post("/v1/projects/{projectID}/secrets", h.CreateSecret)
    body := map[string]any{"key":"K","environment":"stage","value":map[string]any{"x":"y"},"createdBy":"u"}
    b, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/secrets", bytes.NewReader(b))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusBadRequest { t.Fatalf("expected 400, got %d", w.Code) }
}

func TestListSecrets_Middleware_RejectsInvalidEnvQuery(t *testing.T) {
    r := chi.NewRouter()
    r.Use(ValidateEnvQuery([]string{"dev"}))
    h := NewSecretsHandler(&secretsSvcStub{})
    r.Get("/v1/projects/{projectID}/secrets", h.ListSecrets)
    req := httptest.NewRequest(http.MethodGet, "/v1/projects/p1/secrets?env=stage", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusBadRequest { t.Fatalf("expected 400, got %d", w.Code) }
}
