package errors

import (
	"context"
	"math"
	"time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// WithRetry executes fn with exponential backoff retry
func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = fn()
		
		// Success case
		if lastErr == nil {
			return nil
		}

		// Don't retry if error is not retryable
		if !IsRetryable(lastErr) {
			return lastErr
		}

		// Don't sleep after last attempt
		if attempt >= cfg.MaxAttempts {
			break
		}

		// Calculate delay and wait
		delay := exponentialBackoff(attempt, cfg)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// exponentialBackoff calculates delay for an attempt
func exponentialBackoff(attempt int, cfg RetryConfig) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt-1))
	
	if delay > float64(cfg.MaxDelay) {
		return cfg.MaxDelay
	}
	
	return time.Duration(delay)
}
