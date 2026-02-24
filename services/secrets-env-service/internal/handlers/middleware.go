package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"example.com/lma/secrets-env-service/internal/events"
	"example.com/lma/secrets-env-service/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Claims represents JWT claims used by the app
type Claims struct {
	Subject   string
	TenantID  string
	ProjectID string
	Scopes    []string
}

type ctxKeyClaims struct{}

// WithClaims attaches Claims to context for downstream retrieval.
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, ctxKeyClaims{}, claims)
}

// ClaimsFromContext returns claims from context
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	v := ctx.Value(ctxKeyClaims{})
	if v == nil {
		return Claims{}, false
	}
	c, ok := v.(Claims)
	return c, ok
}

// TokenVerifier verifies bearer tokens and returns claims
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

// PanicRecovery recovers from panics and returns 500 JSON
func PanicRecovery(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.SafeLogError(log, fmt.Errorf("panic: %v", rec), "panic recovered")
					Error(w, http.StatusInternalServerError, "internal server error", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// TraceContext ensures a traceparent is present in context (from header or generated).
func TraceContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tp := r.Header.Get("traceparent")
			if tp == "" {
				tp = events.TraceparentFromContext(r.Context())
			}
			ctx := events.WithTraceparent(r.Context(), tp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OtelTracing starts a span per request and annotates basic attributes.
func OtelTracing() func(http.Handler) http.Handler {
	tr := otel.Tracer("http-server")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tr.Start(r.Context(), r.Method+" "+r.URL.Path)
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.target", r.URL.Path),
			)
			defer span.End()
			rw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r.WithContext(ctx))
			span.SetAttributes(attribute.Int("http.status_code", rw.status))
		})
	}
}

// RequestLogger logs basic request information and sets request ID
func RequestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = extractRequestID(r)
				r.Header.Set("X-Request-ID", reqID)
			}
			// Wrap ResponseWriter to capture status
			rw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			log.Info().
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rw.status).
				Dur("duration_ms", time.Since(start)).
				Msg("request")
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// JWTAuth validates bearer token and injects claims into context
func JWTAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				Error(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")
			claims, err := verifier.Verify(r.Context(), token)
			if err != nil || claims == nil {
				Error(w, http.StatusUnauthorized, "invalid token", err)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKeyClaims{}, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScopes ensures caller has all required scopes
func RequireScopes(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				Error(w, http.StatusForbidden, "missing claims", nil)
				return
			}
			scopeSet := map[string]struct{}{}
			for _, s := range claims.Scopes {
				scopeSet[s] = struct{}{}
			}
			for _, req := range required {
				if _, ok := scopeSet[req]; !ok {
					Error(w, http.StatusForbidden, "insufficient scope", nil)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RBACAugment augments scopes based on caller roles (from header X-Roles or existing claims).
// Very simple mapping for demo: admin -> add all write scopes; auditor -> add read scopes.
func RBACAugment() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			rolesHeader := r.Header.Get("X-Roles")
			roles := []string{}
			if rolesHeader != "" {
				roles = strings.Split(rolesHeader, ",")
			}
			aug := claims.Scopes
			for i := range roles {
				roles[i] = strings.TrimSpace(strings.ToLower(roles[i]))
			}
			for _, role := range roles {
				switch role {
				case "admin":
					aug = mergeScopes(aug, []string{"secrets:write", "parity:compute", "leaks:write"})
				case "auditor":
					aug = mergeScopes(aug, []string{"secrets:read", "parity:read", "leaks:read"})
				}
			}
			ctx := context.WithValue(r.Context(), ctxKeyClaims{}, Claims{
				Subject:   claims.Subject,
				TenantID:  claims.TenantID,
				ProjectID: claims.ProjectID,
				Scopes:    aug,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mergeScopes(base, extra []string) []string {
	set := map[string]struct{}{}
	for _, s := range base {
		set[s] = struct{}{}
	}
	for _, s := range extra {
		if _, ok := set[s]; !ok {
			base = append(base, s)
			set[s] = struct{}{}
		}
	}
	return base
}

// ValidateKeyParam ensures chi URLParam("key") is valid if present on the route.
func ValidateKeyParam() func(http.Handler) http.Handler {
	var re = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := chi.URLParam(r, "key")
			if key != "" && !re.MatchString(key) {
				Error(w, http.StatusBadRequest, "invalid key format", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitHTTP enforces a token-bucket limiter using the provided Allow func.
func RateLimitHTTP(allow func(key string) bool) func(http.Handler) http.Handler {
	tr := otel.Tracer("http-server")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "anon"
			if c, ok := ClaimsFromContext(r.Context()); ok {
				key = c.TenantID + ":" + r.Method + ":" + r.URL.Path
			}
			if !allow(key) {
				if span := trace.SpanFromContext(r.Context()); span != nil {
					span.SetAttributes(attribute.String("ratelimit.key", key))
				}
				Error(w, http.StatusTooManyRequests, "rate limit exceeded", nil)
				return
			}
			ctx, span := tr.Start(r.Context(), "ratelimit.allowed")
			defer span.End()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantIsolation compares X-Tenant-ID header with claims.TenantID
func TenantIsolation() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				Error(w, http.StatusForbidden, "missing claims", nil)
				return
			}
			headTenant := r.Header.Get("X-Tenant-ID")
			if headTenant == "" || headTenant != claims.TenantID {
				Error(w, http.StatusForbidden, "tenant mismatch", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// routePattern returns the chi route pattern for the current request, if available.
func routePattern(r *http.Request) string {
	if r == nil {
		return ""
	}
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if p := rctx.RoutePattern(); p != "" {
			return p
		}
	}
	return r.URL.Path
}

// HttpMetrics records HTTP request metrics: counter and latency histogram by method/route/status.
func HttpMetrics(reg prometheus.Registerer) func(http.Handler) http.Handler {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, route, status.",
		},
		[]string{"method", "route", "status"},
	)
	durations := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration by method, route, status.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	reg.MustRegister(requests, durations)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			method := r.Method

			rw := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)

			status := fmt.Sprintf("%d", rw.status)
			route := routePattern(r)

			requests.WithLabelValues(method, route, status).Inc()
			durations.WithLabelValues(method, route, status).Observe(time.Since(start).Seconds())
		})
	}
}
