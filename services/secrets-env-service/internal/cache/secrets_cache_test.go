package cache

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	miniredis "github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
)

type fakeRepo struct{ s *domain.Secret; list []*domain.Secret }

func (f *fakeRepo) Create(ctx context.Context, s *domain.Secret) error { f.s = s; return nil }
func (f *fakeRepo) GetByKey(ctx context.Context, projectID, key, environment string) (*domain.Secret, error) { return f.s, nil }
func (f *fakeRepo) List(ctx context.Context, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) { return f.list, "", nil }
func (f *fakeRepo) Update(ctx context.Context, s *domain.Secret) error { f.s = s; return nil }
func (f *fakeRepo) Delete(ctx context.Context, id string) error { return nil }

func TestSecretsCache_GetByKey_CachesAndTTL(t *testing.T) {
	mr, _ := miniredis.Run(); defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	base := &fakeRepo{s: &domain.Secret{ID:"1", ProjectID:"p", Environment:"dev", Key:"K"}}
	c := NewSecretsCacheRepository(base, rdb, time.Second)
	ctx := context.Background()
	// 1st: miss -> base
	if _, err := c.GetByKey(ctx, "p", "K", "dev"); err != nil { t.Fatalf("err: %v", err) }
	// 2nd: hit cache
	if _, err := c.GetByKey(ctx, "p", "K", "dev"); err != nil { t.Fatalf("err: %v", err) }
	// ensure ttl roughly set (>0)
	if ttl := mr.TTL("cache:secret:project:p:env:dev:key:K"); ttl <= 0 { t.Fatalf("expected ttl >0, got %v", ttl) }
}

func TestSecretsCache_List_CachesPerPage(t *testing.T) {
	mr, _ := miniredis.Run(); defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	base := &fakeRepo{list: []*domain.Secret{{ID:"1"},{ID:"2"}}}
	c := NewSecretsCacheRepository(base, rdb, time.Minute)
	ctx := context.Background()
	_, _, _ = c.List(ctx, "p", "dev", 2, "")
	if !mr.Exists("cache:secret:project:p:env:dev:list:2:-") { t.Fatalf("missing list cache") }
}
