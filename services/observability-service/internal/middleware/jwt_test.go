package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeVerifier struct{ out *Claims; err error }
func (f *fakeVerifier) Verify(ctx context.Context, token string) (*Claims, error){ if f.err!=nil { return nil, f.err }; return f.out, nil }

func TestJWTMiddleware_ValidToken_PassesThrough(t *testing.T){
	fv := &fakeVerifier{ out: &Claims{ Subject:"sub", Scopes: []string{"observability:read"} } }
	mw := JWT(fv)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ called = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	if !called { t.Fatalf("expected next to be called") }
	if rec.Code != 200 { t.Fatalf("unexpected status: %d", rec.Code) }
}

func TestJWTMiddleware_MissingToken_401(t *testing.T){
	mw := JWT(&fakeVerifier{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter,*http.Request){})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("want 401 got %d", rec.Code) }
}

func TestJWTMiddleware_InvalidSignature_401(t *testing.T){
	mw := JWT(&fakeVerifier{ err: errors.New("bad")})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(http.ResponseWriter,*http.Request){})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("want 401 got %d", rec.Code) }
}
