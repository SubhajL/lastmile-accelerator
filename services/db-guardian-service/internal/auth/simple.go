package auth

import (
	"context"
	"fmt"
	"strings"
)

// SimpleAuthenticator accepts any non-empty Bearer token and does no scope enforcement.
// Suitable only for dev/local until a real JWT/JWKS authenticator is wired.
type SimpleAuthenticator struct{}

func NewSimpleAuthenticator() *SimpleAuthenticator { return &SimpleAuthenticator{} }

func (a *SimpleAuthenticator) Verify(_ context.Context, bearer string, required []string) (*Claims, error) {
	if !strings.HasPrefix(bearer, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header")
	}
	tok := strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))
	if tok == "" {
		return nil, fmt.Errorf("empty token")
	}
	// No scope enforcement here; callers should provide a resolver and a real authenticator in prod.
	return &Claims{Subject: "bearer", Scopes: []string{}}, nil
}