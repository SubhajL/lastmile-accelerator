package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSecretsService struct {
	created *domain.Secret
	value   map[string]interface{}
	list    []*domain.Secret
}

func (f *fakeSecretsService) CreateSecret(_ any, s *domain.Secret, v map[string]interface{}) error {
	f.created = s
	f.value = v
	return nil
}
func (f *fakeSecretsService) GetSecret(_ any, tenantID, projectID, key, env string) (*domain.Secret, map[string]interface{}, error) {
	return &domain.Secret{ID: "id1", TenantID: tenantID, ProjectID: projectID, Key: key, Environment: env, CreatedAt: time.Now()}, map[string]interface{}{"k":"v"}, nil
}
func (f *fakeSecretsService) DeleteSecret(_ any, _ string, _ string, _ string, _ string) error { return nil }
func (f *fakeSecretsService) ListSecrets(_ any, _ string, _ string, _ int, _ string) ([]*domain.Secret, string, error) {
	if f.list == nil {
		f.list = []*domain.Secret{{ID: "a"}, {ID: "b"}}
	}
	return f.list, "", nil
}

func TestSecretsHandler_Create_ValidRequest(t *testing.T) {
	svc := &fakeSecretsService{}
	h := NewSecretsHandler(svc)

	r := chi.NewRouter()
	r.Post("/v1/projects/{projectID}/secrets", h.CreateSecret)

	body := map[string]any{"key":"API_KEY","environment":"production","value":map[string]any{"api_key":"x"},"createdBy":"user@example.com"}
	b,_ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/secrets", bytes.NewReader(b))
	req.Header.Set("Content-Type","application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, svc.created)
	assert.Equal(t, "API_KEY", svc.created.Key)
	assert.Equal(t, "production", svc.created.Environment)
}

func TestSecretsHandler_GetSecret_Existing(t *testing.T) {
	svc := &fakeSecretsService{}
	h := NewSecretsHandler(svc)
	r := chi.NewRouter()
	r.Get("/v1/projects/{projectID}/secrets/{key}", h.GetSecret)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/proj-1/secrets/API_KEY?env=staging", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

func TestSecretsHandler_Delete_Valid(t *testing.T) {
	svc := &fakeSecretsService{}
	h := NewSecretsHandler(svc)
	r := chi.NewRouter()
	r.Delete("/v1/projects/{projectID}/secrets/{key}", h.DeleteSecret)

	req := httptest.NewRequest(http.MethodDelete, "/v1/projects/proj-1/secrets/API_KEY?env=dev", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
}

func TestSecretsHandler_List_Paginated(t *testing.T) {
	svc := &fakeSecretsService{}
	h := NewSecretsHandler(svc)
	r := chi.NewRouter()
	r.Get("/v1/projects/{projectID}/secrets", h.ListSecrets)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/proj-1/secrets?env=prod&limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
