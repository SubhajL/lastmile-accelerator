package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/google/uuid"
)

// SecretsRepository handles secret metadata persistence
type SecretsRepository struct {
	db *sql.DB

	// Test mode fields
	testMode bool
	testData map[string]*domain.Secret
	mu       sync.RWMutex
}

// NewSecretsRepository creates a new repository
func NewSecretsRepository(db *sql.DB) *SecretsRepository {
	if db == nil {
		return &SecretsRepository{
			testMode: true,
			testData: make(map[string]*domain.Secret),
		}
	}
	return &SecretsRepository{
		db: db,
	}
}

// Create inserts a new secret metadata record
func (r *SecretsRepository) Create(ctx context.Context, secret *domain.Secret) error {
	if r.testMode {
		r.mu.Lock()
		defer r.mu.Unlock()

		// Check for duplicates (project+key+env)
		for _, existing := range r.testData {
			if existing.ProjectID == secret.ProjectID &&
				existing.Key == secret.Key &&
				existing.Environment == secret.Environment {
				return fmt.Errorf("secret already exists: %s in %s environment", secret.Key, secret.Environment)
			}
		}

		// Store a copy to avoid pointer issues
		copy := *secret
		r.testData[secret.ID] = &copy
		return nil
	}

	query := `
		INSERT INTO secrets (id, tenant_id, project_id, key, environment, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		secret.ID,
		secret.TenantID,
		secret.ProjectID,
		secret.Key,
		secret.Environment,
		secret.CreatedAt,
		secret.UpdatedAt,
		secret.CreatedBy,
	)

	return err
}

// GetByID retrieves a secret by its UUID
func (r *SecretsRepository) GetByID(ctx context.Context, id string) (*domain.Secret, error) {
	if r.testMode {
		r.mu.RLock()
		defer r.mu.RUnlock()

		secret, exists := r.testData[id]
		if !exists {
			return nil, fmt.Errorf("secret not found: %s", id)
		}
		return secret, nil
	}

	query := `
		SELECT id, tenant_id, project_id, key, environment, created_at, updated_at, created_by
		FROM secrets
		WHERE id = $1
	`

	secret := &domain.Secret{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&secret.ID,
		&secret.TenantID,
		&secret.ProjectID,
		&secret.Key,
		&secret.Environment,
		&secret.CreatedAt,
		&secret.UpdatedAt,
		&secret.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("secret not found: %s", id)
	}

	return secret, err
}

// GetByKey retrieves a secret by natural key (project+key+env)
func (r *SecretsRepository) GetByKey(ctx context.Context, projectID, key, environment string) (*domain.Secret, error) {
	if r.testMode {
		r.mu.RLock()
		defer r.mu.RUnlock()

		for _, secret := range r.testData {
			if secret.ProjectID == projectID &&
				secret.Key == key &&
				secret.Environment == environment {
				return secret, nil
			}
		}
		return nil, fmt.Errorf("secret not found: %s/%s/%s", projectID, key, environment)
	}

	query := `
		SELECT id, tenant_id, project_id, key, environment, created_at, updated_at, created_by
		FROM secrets
		WHERE project_id = $1 AND key = $2 AND environment = $3
	`

	secret := &domain.Secret{}
	err := r.db.QueryRowContext(ctx, query, projectID, key, environment).Scan(
		&secret.ID,
		&secret.TenantID,
		&secret.ProjectID,
		&secret.Key,
		&secret.Environment,
		&secret.CreatedAt,
		&secret.UpdatedAt,
		&secret.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("secret not found: %s/%s/%s", projectID, key, environment)
	}

	return secret, err
}

// Update modifies secret metadata
func (r *SecretsRepository) Update(ctx context.Context, secret *domain.Secret) error {
	if r.testMode {
		r.mu.Lock()
		defer r.mu.Unlock()

		_, exists := r.testData[secret.ID]
		if !exists {
			return fmt.Errorf("secret not found: %s", secret.ID)
		}

		// Update only the UpdatedAt field
		r.testData[secret.ID].UpdatedAt = secret.UpdatedAt
		return nil
	}

	query := `
		UPDATE secrets
		SET updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, secret.UpdatedAt, secret.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("secret not found: %s", secret.ID)
	}

	return nil
}

// Delete removes a secret metadata record
func (r *SecretsRepository) Delete(ctx context.Context, id string) error {
	if r.testMode {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.testData, id)
		return nil
	}

	query := `DELETE FROM secrets WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List returns paginated secrets filtered by project and environment
func (r *SecretsRepository) List(ctx context.Context, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
	if r.testMode {
		r.mu.RLock()
		defer r.mu.RUnlock()

		// Filter by project and environment
		var filtered []*domain.Secret
		for _, secret := range r.testData {
			if secret.ProjectID == projectID && secret.Environment == environment {
				filtered = append(filtered, secret)
			}
		}

		// Sort by created_at DESC
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		})

		// Simple pagination (skip based on cursor)
		start := 0
		if cursor != "" {
			// In test mode, cursor is just an index
            _, _ = fmt.Sscanf(cursor, "%d", &start)
		}

		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}

		var result []*domain.Secret
		nextCursor := ""
		
		if start < len(filtered) {
			result = filtered[start:end]
			if end < len(filtered) {
				nextCursor = fmt.Sprintf("%d", end)
			}
		}

		return result, nextCursor, nil
	}

	query := `
		SELECT id, tenant_id, project_id, key, environment, created_at, updated_at, created_by
		FROM secrets
		WHERE project_id = $1 AND environment = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

    rows, err := r.db.QueryContext(ctx, query, projectID, environment, limit)
	if err != nil {
		return nil, "", err
	}
    defer func(){ _ = rows.Close() }()

	var secrets []*domain.Secret
	for rows.Next() {
		secret := &domain.Secret{}
		err := rows.Scan(
			&secret.ID,
			&secret.TenantID,
			&secret.ProjectID,
			&secret.Key,
			&secret.Environment,
			&secret.CreatedAt,
			&secret.UpdatedAt,
			&secret.CreatedBy,
		)
		if err != nil {
			return nil, "", err
		}
		secrets = append(secrets, secret)
	}

	nextCursor := ""
	if len(secrets) == limit {
		// Use last ID as cursor
		nextCursor = secrets[len(secrets)-1].ID
	}

	return secrets, nextCursor, nil
}

// CreateVersion records a version history entry
func (r *SecretsRepository) CreateVersion(ctx context.Context, version *domain.SecretVersion) error {
	if r.testMode {
		// In test mode, we don't track versions
		return nil
	}

	if version.ID == "" {
		version.ID = uuid.New().String()
	}

	query := `
		INSERT INTO secret_versions (id, secret_id, version_number, created_at, created_by, rotated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		version.ID,
		version.SecretID,
		version.VersionNumber,
		version.CreatedAt,
		version.CreatedBy,
		version.RotatedAt,
	)

	return err
}
