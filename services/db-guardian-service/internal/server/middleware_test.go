package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONValidation_RejectsWrongContentType(t *testing.T) {
	mw := JSONValidation(1024)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := mw(next)
	r := httptest.NewRequest(http.MethodPost, "/x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnsupportedMediaType { t.Fatalf("expected 415, got %d", w.Code) }
}

type fakeAuth struct{ ok bool }
func (a *fakeAuth) Verify(_ context.Context, _ string, _ []string) (*Claims, error) { if a.ok { return &Claims{Subject:"u"}, nil }; return nil, fmt.Errorf("bad") }

type denyLimiter struct{}
func (d *denyLimiter) Allow(_ context.Context, _ string, _ int) (bool, error) { return false, nil }

func TestRequireScopes_AuthZ(t *testing.T) {
	mw := RequireScopes(&fakeAuth{ok:true}, "s1")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := mw(next)
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("Authorization","Bearer x")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestRateLimit_Denies(t *testing.T) {
	mw := RateLimit(&denyLimiter{}, func(r *http.Request) string { return "key" }, 1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := mw(next)
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests { t.Fatalf("expected 429, got %d", w.Code) }
}
