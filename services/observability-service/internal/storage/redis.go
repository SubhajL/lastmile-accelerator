package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps redis.Client with helper methods
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates Redis client with connection pooling
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// HealthCheck executes PING command with 2s timeout
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	status := r.client.Ping(ctx)
	if err := status.Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes Redis connection
func (r *RedisClient) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}

// Client returns underlying redis.Client
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Name returns the dependency name for health checks
func (r *RedisClient) Name() string {
	return "redis"
}
