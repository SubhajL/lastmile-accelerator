package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
)

func TestRBAC_Admin_AllowsWriteEndpoints(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RBACAugment())
    // Protect with write scope
    r.Post("/v1/projects/{projectID}/secrets", RequireScopes("secrets:write")(http.HandlerFunc(okHandler)).ServeHTTP)
    req := httptest.NewRequest(http.MethodPost, "/v1/projects/p/secrets", nil)
    claims := Claims{TenantID: "t1", ProjectID: "p", Roles: []string{"admin"}}
    req = req.WithContext(WithClaims(req.Context(), claims))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestRBAC_Auditor_ReadOnly(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RBACAugment())
    // GET protected by read scope
    r.Get("/v1/projects/{projectID}/secrets", RequireScopes("secrets:read")(http.HandlerFunc(okHandler)).ServeHTTP)
    // POST protected by write scope
    r.Post("/v1/projects/{projectID}/secrets", RequireScopes("secrets:write")(http.HandlerFunc(okHandler)).ServeHTTP)

    // Read OK
    getReq := httptest.NewRequest(http.MethodGet, "/v1/projects/p/secrets", nil)
    getReq = getReq.WithContext(WithClaims(getReq.Context(), Claims{TenantID:"t1", ProjectID:"p", Roles:[]string{"auditor"}}))
    gw := httptest.NewRecorder()
    r.ServeHTTP(gw, getReq)
    if gw.Code != 200 { t.Fatalf("expected 200, got %d", gw.Code) }

    // Write denied
    postReq := httptest.NewRequest(http.MethodPost, "/v1/projects/p/secrets", nil)
    postReq = postReq.WithContext(WithClaims(postReq.Context(), Claims{TenantID:"t1", ProjectID:"p", Roles:[]string{"auditor"}}))
    pw := httptest.NewRecorder()
    r.ServeHTTP(pw, postReq)
    if pw.Code != http.StatusForbidden { t.Fatalf("expected 403, got %d", pw.Code) }
}

func TestRBAC_NoRoles_Denied(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RBACAugment())
    r.Post("/v1/projects/{projectID}/secrets", RequireScopes("secrets:write")(http.HandlerFunc(okHandler)).ServeHTTP)
    req := httptest.NewRequest(http.MethodPost, "/v1/projects/p/secrets", nil)
    req = req.WithContext(WithClaims(req.Context(), Claims{TenantID:"t1", ProjectID:"p"}))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403, got %d", w.Code) }
}
