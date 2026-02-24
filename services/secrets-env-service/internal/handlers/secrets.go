package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type secretsServicePort interface {
	CreateSecret(ctx any, secret *domain.Secret, value map[string]interface{}) error
	GetSecret(ctx any, tenantID, projectID, key, environment string) (*domain.Secret, map[string]interface{}, error)
	DeleteSecret(ctx any, tenantID, projectID, key, environment string) error
	ListSecrets(ctx any, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error)
}

type SecretsHandler struct {
	svc secretsServicePort
	envAllowlist EnvAllowlist
}

func NewSecretsHandler(svc secretsServicePort, envAllowlist EnvAllowlist) *SecretsHandler {
	return &SecretsHandler{svc: svc, envAllowlist: envAllowlist}
}

type createSecretRequest struct {
	Key         string                 `json:"key"`
	Environment string                 `json:"environment"`
	Value       map[string]interface{} `json:"value"`
	CreatedBy   string                 `json:"createdBy"`
}

type secretMeta struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (h *SecretsHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	var req createSecretRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid json", err)
		return
	}
	fields := map[string]string{}
	if req.Key == "" { fields["key"] = "required" }
	if req.Environment == "" { fields["environment"] = "required" }
	// key format validation
	var re = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	if req.Key != "" && !re.MatchString(req.Key) { fields["key"] = "invalid format" }
	if req.Environment != "" && !h.envAllowlist.Allows(req.Environment) { fields["environment"] = "not allowed" }
	if len(fields) > 0 { ValidationError(w, fields); return }

	id := uuid.New().String()
	claims, _ := ClaimsFromContext(r.Context())
	secret := &domain.Secret{
		ID:          id,
		TenantID:    claims.TenantID,
		ProjectID:   projectID,
		Key:         req.Key,
		Environment: req.Environment,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   req.CreatedBy,
	}
	if err := h.svc.CreateSecret(r.Context(), secret, req.Value); err != nil {
		Error(w, http.StatusInternalServerError, "failed to create secret", err)
		return
	}
	w.Header().Set("Location", r.URL.Path+"/"+req.Key)
	Success(w, http.StatusCreated, secretMeta{ID: id, Key: secret.Key, Environment: secret.Environment, CreatedAt: secret.CreatedAt})
}

func (h *SecretsHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	key := chi.URLParam(r, "key")
	env := r.URL.Query().Get("env")
	if !h.envAllowlist.AllowsOptional(env) { ValidationError(w, map[string]string{"env":"not allowed"}); return }
	claims, _ := ClaimsFromContext(r.Context())
	secret, _, err := h.svc.GetSecret(r.Context(), claims.TenantID, projectID, key, env)
	if err != nil {
		Error(w, http.StatusNotFound, "secret not found", err)
		return
	}
	Success(w, http.StatusOK, secretMeta{ID: secret.ID, Key: secret.Key, Environment: secret.Environment, CreatedAt: secret.CreatedAt})
}

func (h *SecretsHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	key := chi.URLParam(r, "key")
	env := r.URL.Query().Get("env")
	if !h.envAllowlist.AllowsOptional(env) { ValidationError(w, map[string]string{"env":"not allowed"}); return }
	claims, _ := ClaimsFromContext(r.Context())
	if err := h.svc.DeleteSecret(r.Context(), claims.TenantID, projectID, key, env); err != nil {
		Error(w, http.StatusNotFound, "secret not found", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SecretsHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	env := r.URL.Query().Get("env")
	if !h.envAllowlist.AllowsOptional(env) { ValidationError(w, map[string]string{"env":"not allowed"}); return }
	limitStr := r.URL.Query().Get("limit")
	cursor := r.URL.Query().Get("cursor")
	limit := 50
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil { limit = n }
	}
	list, next, err := h.svc.ListSecrets(r.Context(), projectID, env, limit, cursor)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list secrets", err)
		return
	}
	// convert to meta slice
	out := make([]secretMeta, 0, len(list))
	for _, s := range list {
		out = append(out, secretMeta{ID: s.ID, Key: s.Key, Environment: s.Environment, CreatedAt: s.CreatedAt})
	}
	Success(w, http.StatusOK, map[string]any{"items": out, "next": next})
}
