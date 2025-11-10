package repository

import (
	"context"
	"example.com/lma/secrets-env-service/internal/domain"
)

// SecretsMetaRepo abstracts metadata persistence for secrets (not values).
type SecretsMetaRepo interface {
	Create(ctx context.Context, secret *domain.Secret) error
	GetByKey(ctx context.Context, projectID, key, environment string) (*domain.Secret, error)
	List(ctx context.Context, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error)
	Update(ctx context.Context, secret *domain.Secret) error
	Delete(ctx context.Context, id string) error
}
