package cache

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil { return nil, err }
	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil { return nil, err }
	return rdb, nil
}
