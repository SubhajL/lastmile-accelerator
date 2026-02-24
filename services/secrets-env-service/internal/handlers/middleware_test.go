package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appLogger "example.com/lma/secrets-env-service/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type fakeVerifier struct{ allow bool; claims Claims; err error }

func (f *fakeVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	if !f.allow {
		return nil, f.err
	}
	return &f.claims, nil
}

func TestPanicRecovery_Returns500JSON(t *testing.T) {
	var buf bytes.Buffer
	log := appLogger.New("test", "info", &buf)

	rec := PanicRecovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Error)
}

func TestRequestLogger_SetsRequestIDAndLogs(t *testing.T) {
	var buf bytes.Buffer
	log := appLogger.New("test", "info", &buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	rec := RequestLogger(log)(next)

	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)

	assert.Equal(t, 204, w.Code)
	// log line should contain service and request fields
	assert.Contains(t, buf.String(), "\"service\":\"test\"")
}

func TestJWTAuth_ValidTokenSetsClaims(t *testing.T) {
	ver := &fakeVerifier{allow: true, claims: Claims{
		Subject:  "user-1",
		TenantID: "tenant-1",
		ProjectID: "proj-1",
		Scopes:   []string{"secrets:read"},
	}}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := ClaimsFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "tenant-1", c.TenantID)
		w.WriteHeader(200)
	})

	rec := JWTAuth(ver)(next)
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestJWTAuth_MissingTokenReturns401(t *testing.T) {
	rec := JWTAuth(&fakeVerifier{allow: true})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireScopes_ForbiddenWhenMissing(t *testing.T) {
	rec := RequireScopes("secrets:write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	// context without scopes
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTenantIsolation_HeaderMismatch(t *testing.T) {
	// set claims in context
	next := TenantIsolation()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	// mismatch header
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("X-Tenant-ID", "other-tenant")
	ctx := context.WithValue(r.Context(), ctxKeyClaims{}, Claims{TenantID: "tenant-1"})
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r.WithContext(ctx))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTenantIsolation_HeaderMatch(t *testing.T) {
	next := TenantIsolation()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("X-Tenant-ID", "tenant-1")
	ctx := context.WithValue(r.Context(), ctxKeyClaims{}, Claims{TenantID: "tenant-1"})
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r.WithContext(ctx))
	assert.Equal(t, 200, w.Code)
}

func TestValidateKeyParam_Invalid(t *testing.T) {
	next := ValidateKeyParam()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
r := httptest.NewRequest(http.MethodGet, "/v1/projects/p/secrets/valid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "INVALID KEY")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequireJSONContentType_AllowsApplicationJSON(t *testing.T) {
	next := RequireJSONContentType()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("{}")))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	assert.Equal(t, 204, w.Code)
}

func TestRequireJSONContentType_RejectsMissingContentType(t *testing.T) {
	next := RequireJSONContentType()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)
}

// Ensure zerolog imported for compilation in tests
var _ zerolog.Logger
