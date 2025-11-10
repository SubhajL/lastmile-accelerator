package middleware

import (
	"context"
	"net/http"
)

// Claims represents a reduced set of JWT claims we care about.
type Claims struct {
	Subject   string
	ProjectID string
	Scopes    []string
	Raw       map[string]interface{}
}

// Verifier abstracts token verification for testability.
type Verifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

// claimsKey is the context key for Claims.
type claimsKey struct{}

// WithClaims returns a context carrying Claims.
func WithClaims(ctx context.Context, c *Claims) context.Context { return context.WithValue(ctx, claimsKey{}, c) }

// FromClaims fetches Claims from context.
func FromClaims(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	return c, ok
}

// JWT returns a middleware that validates Authorization: Bearer tokens via provided verifier.
func JWT(verifier Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if len(auth) < 8 || auth[:7] != "Bearer " {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tok := auth[7:]
			claims, err := verifier.Verify(r.Context(), tok)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}
