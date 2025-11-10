package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireScopes_MissingClaims_403(t *testing.T){
	mw := RequireScopes("observability:read")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(http.HandlerFunc(func(http.ResponseWriter,*http.Request){})).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden { t.Fatalf("want 403 got %d", rec.Code) }
}

func TestRequireScopes_Present_200(t *testing.T){
	mw := RequireScopes("observability:write")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	claims := &Claims{ Scopes: []string{"observability:write"} }
	next := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200) }))
	next.ServeHTTP(rec, req.WithContext(WithClaims(req.Context(), claims)))
	if rec.Code != 200 { t.Fatalf("want 200 got %d", rec.Code) }
}
