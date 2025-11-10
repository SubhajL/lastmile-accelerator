package ratelimit

import (
	"context"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, cost int) (bool, error)
}
