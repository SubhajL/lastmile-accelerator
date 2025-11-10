package auth

import (
	"context"
)

type Claims struct {
	Subject string
	Scopes  []string
}

type Authenticator interface {
	Verify(ctx context.Context, bearer string, required []string) (*Claims, error)
}
