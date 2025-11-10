package middleware

import (
	"net/http"
	"slices"
)

// RequireScopes ensures the JWT claims contain all required scopes.
func RequireScopes(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := FromClaims(r.Context())
			if !ok {
				http.Error(w, "missing claims", http.StatusForbidden)
				return
			}
			for _, need := range required {
				if !slices.Contains(claims.Scopes, need) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
