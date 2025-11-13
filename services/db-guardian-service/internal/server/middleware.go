package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// RequireScopesFunc resolves required scopes per request
func RequireScopesFunc(authn Authenticator, resolve func(*http.Request) []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authn == nil { next.ServeHTTP(w, r); return }
            required := []string{}
            if resolve != nil { required = resolve(r) }
            if required == nil { // explicit bypass
                next.ServeHTTP(w, r); return
            }
			h := r.Header.Get("Authorization")
			if h == "" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
			if _, err := authn.Verify(r.Context(), h, required); err != nil {
				http.Error(w, "forbidden", http.StatusForbidden); return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var (
	requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"}, []string{"method","path","code"})
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request duration seconds", Buckets: prometheus.DefBuckets}, []string{"method","path"})
)

func init() { prometheus.MustRegister(requestCounter, requestLatency) }

// Metrics records per-request count and latency
func Metrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, code: 200}
			next.ServeHTTP(rw, r)
			dur := time.Since(start).Seconds()
			path := r.URL.Path
			requestCounter.WithLabelValues(r.Method, path, http.StatusText(rw.code)).Inc()
			requestLatency.WithLabelValues(r.Method, path).Observe(dur)
		})
	}
}

type statusRecorder struct { http.ResponseWriter; code int }
func (sr *statusRecorder) WriteHeader(code int) { sr.code = code; sr.ResponseWriter.WriteHeader(code) }

// JSONValidation ensures proper JSON content-type and limits body size
func JSONValidation(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.Header.Get("Content-Type") != "application/json" {
					http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType); return
				}
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireScopes verifies JWT via deps.Authenticator; if deps.Authenticator is nil, it is a no-op
func RequireScopes(authn Authenticator, scopes ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authn == nil { next.ServeHTTP(w, r); return }
			h := r.Header.Get("Authorization")
			if h == "" { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
			if _, err := authn.Verify(r.Context(), h, scopes); err != nil {
				http.Error(w, "forbidden", http.StatusForbidden); return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit enforces rate limit; if limiter is nil, it is a no-op
func RateLimit(limiter RateLimiter, keyFn func(*http.Request) string, cost int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil { next.ServeHTTP(w, r); return }
			key := keyFn(r)
			ok, _ := limiter.Allow(r.Context(), key, cost)
			if !ok { http.Error(w, "rate limit exceeded", http.StatusTooManyRequests); return }
			next.ServeHTTP(w, r)
		})
	}
}
