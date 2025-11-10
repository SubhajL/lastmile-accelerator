package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithRetry_SuccessFirstAttempt(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return nil
	}

	err := WithRetry(context.Background(), cfg, fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return NewAPIError(503, "service unavailable", nil)
		}
		return nil
	}

	err := WithRetry(context.Background(), cfg, fn)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestWithRetry_MaxAttemptsExceeded(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return NewAPIError(503, "service unavailable", nil)
	}

	err := WithRetry(context.Background(), cfg, fn)
	assert.Error(t, err)
	assert.Equal(t, 2, attempts)
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	attempts := 0
	fn := func() error {
		attempts++
		if attempts == 2 {
			cancel()
		}
		return NewAPIError(503, "service unavailable", nil)
	}

	err := WithRetry(ctx, cfg, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
	assert.LessOrEqual(t, attempts, 3) // Should stop after context cancelled
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return NewAPIError(400, "bad request", nil)
	}

	err := WithRetry(context.Background(), cfg, fn)
	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry
}

func TestExponentialBackoff_IncreasesDelay(t *testing.T) {
	cfg := RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	delay1 := exponentialBackoff(1, cfg)
	delay2 := exponentialBackoff(2, cfg)
	delay3 := exponentialBackoff(3, cfg)

	assert.Equal(t, 100*time.Millisecond, delay1)
	assert.Equal(t, 200*time.Millisecond, delay2)
	assert.Equal(t, 400*time.Millisecond, delay3)
}

func TestExponentialBackoff_CappedAtMax(t *testing.T) {
	cfg := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
	}

	delay10 := exponentialBackoff(10, cfg)
	assert.Equal(t, 2*time.Second, delay10)
}
