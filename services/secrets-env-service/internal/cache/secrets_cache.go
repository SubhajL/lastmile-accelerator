package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/repository"
	redis "github.com/redis/go-redis/v9"
)

type SecretsCacheRepository struct {
	base repository.SecretsMetaRepo
	rdb  *redis.Client
	ttl  time.Duration
}

func NewSecretsCacheRepository(base repository.SecretsMetaRepo, rdb *redis.Client, ttl time.Duration) *SecretsCacheRepository {
	return &SecretsCacheRepository{base: base, rdb: rdb, ttl: ttl}
}

func keyGet(projectID, env, key string) string {
	return fmt.Sprintf("cache:secret:project:%s:env:%s:key:%s", projectID, env, key)
}
func keyList(projectID, env string, limit int, cursor string) string {
	if cursor == "" { cursor = "-" }
	return fmt.Sprintf("cache:secret:project:%s:env:%s:list:%d:%s", projectID, env, limit, cursor)
}

func (c *SecretsCacheRepository) Create(ctx context.Context, s *domain.Secret) error {
	if err := c.base.Create(ctx, s); err != nil { return err }
	_ = c.rdb.Del(ctx, keyGet(s.ProjectID, s.Environment, s.Key)).Err()
	return nil
}

func (c *SecretsCacheRepository) GetByKey(ctx context.Context, projectID, key, environment string) (*domain.Secret, error) {
	k := keyGet(projectID, environment, key)
	if b, err := c.rdb.Get(ctx, k).Bytes(); err == nil {
		var s domain.Secret
		if json.Unmarshal(b, &s) == nil { return &s, nil }
	}
	s, err := c.base.GetByKey(ctx, projectID, key, environment)
	if err != nil { return nil, err }
	if b, err := json.Marshal(s); err == nil { _ = c.rdb.Set(ctx, k, b, c.ttl).Err() }
	return s, nil
}

func (c *SecretsCacheRepository) List(ctx context.Context, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
	k := keyList(projectID, environment, limit, cursor)
	if b, err := c.rdb.Get(ctx, k).Bytes(); err == nil {
		var wrap struct{ Items []*domain.Secret; Next string }
		if json.Unmarshal(b, &wrap) == nil { return wrap.Items, wrap.Next, nil }
	}
	items, next, err := c.base.List(ctx, projectID, environment, limit, cursor)
	if err != nil { return nil, "", err }
	if b, err := json.Marshal(struct{ Items []*domain.Secret; Next string }{Items: items, Next: next}); err == nil { _ = c.rdb.Set(ctx, k, b, c.ttl).Err() }
	return items, next, nil
}

func (c *SecretsCacheRepository) Update(ctx context.Context, s *domain.Secret) error {
	if err := c.base.Update(ctx, s); err != nil { return err }
	_ = c.rdb.Del(ctx, keyGet(s.ProjectID, s.Environment, s.Key)).Err()
	return nil
}

func (c *SecretsCacheRepository) Delete(ctx context.Context, id string) error {
	// cannot know composite key; rely on downstream to call Get before Delete if needed; delete anyway
	return c.base.Delete(ctx, id)
}
