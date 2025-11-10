package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handlers "example.com/lma/secrets-env-service/internal/handlers"
	"example.com/lma/secrets-env-service/internal/security"
)

func TestRateLimitHTTP_TooManyRequests(t *testing.T) {
	lim := security.NewRateLimiter(1, 1) // 1 rps, burst=1
	next := handlers.RateLimitHTTP(lim.Allow)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w1 := httptest.NewRecorder()
	next.ServeHTTP(w1, r)
	if w1.Code != 200 { t.Fatalf("first request should pass, got %d", w1.Code) }

	w2 := httptest.NewRecorder()
	next.ServeHTTP(w2, r)
	if w2.Code != http.StatusTooManyRequests { t.Fatalf("second request should be 429, got %d", w2.Code) }
}
