package service

import (
	"context"
	"fmt"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/events"
    "example.com/lma/secrets-env-service/internal/metrics"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/vault"
)

// SecretsService orchestrates secret operations
type SecretsService struct {
	vault *vault.Client
	repo  repository.SecretsMetaRepo
	audit auditWriter
	pub   events.Publisher
}

type auditWriter interface { Write(ctx context.Context, e *domain.AuditLogEntry) error }

// NewSecretsService creates a new secrets service
func NewSecretsService(vault *vault.Client, repo repository.SecretsMetaRepo, audit auditWriter, pub events.Publisher) *SecretsService {
	return &SecretsService{
		vault: vault,
		repo:  repo,
		audit: audit,
		pub:   pub,
	}
}

// CreateSecret creates a new secret with value stored in Vault
func (s *SecretsService) CreateSecret(ctx context.Context, secret *domain.Secret, value map[string]interface{}) error {
	// Validate secret metadata
	if err := secret.Validate(); err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("create", "error") }
		return fmt.Errorf("validation failed: %w", err)
	}

	// Write value to Vault first
	vaultPath := secret.VaultPath()
    if err := s.vault.WriteSecret(ctx, vaultPath, value); err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("create", "error") }
		return fmt.Errorf("failed to write to vault: %w", err)
	}

	// Save metadata to database
    if err := s.repo.Create(ctx, secret); err != nil {
        // Attempt rollback (delete from Vault)
        _ = s.vault.DeleteSecret(ctx, vaultPath)
        if metrics.Default != nil { metrics.Default.ObserveSecrets("create", "error") }
		return fmt.Errorf("failed to save metadata: %w", err)
	}

// Audit & publish
if s.audit != nil { _ = s.audit.Write(ctx, &domain.AuditLogEntry{TenantID: secret.TenantID, ProjectID: secret.ProjectID, Key: secret.Key, Environment: secret.Environment, Action: "created", Actor: secret.CreatedBy, OccurredAt: secret.CreatedAt}) }
s.publishEvent("secret.created", map[string]interface{}{
		"secret_id":  secret.ID,
		"project_id": secret.ProjectID,
		"key":        secret.Key,
		"environment": secret.Environment,
})
    if metrics.Default != nil { metrics.Default.ObserveSecrets("create", "success") }

	return nil
}

// GetSecret retrieves a secret with its value
func (s *SecretsService) GetSecret(ctx context.Context, tenantID, projectID, key, environment string) (*domain.Secret, map[string]interface{}, error) {
	// Get metadata from database
	secret, err := s.repo.GetByKey(ctx, projectID, key, environment)
	if err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("get", "error") }
		return nil, nil, err
	}

	// Verify tenant isolation
	if secret.TenantID != tenantID {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("get", "error") }
		return nil, nil, fmt.Errorf("tenant mismatch: unauthorized access")
	}

	// Get value from Vault
	vaultPath := secret.VaultPath()
    value, err := s.vault.ReadSecret(ctx, vaultPath)
    if err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("get", "error") }
		return nil, nil, fmt.Errorf("failed to read from vault: %w", err)
	}

// Audit & publish access
if s.audit != nil { _ = s.audit.Write(ctx, &domain.AuditLogEntry{TenantID: secret.TenantID, ProjectID: projectID, Key: key, Environment: environment, Action: "accessed", Actor: "", OccurredAt: secret.UpdatedAt}) }
s.publishEvent("secret.accessed", map[string]interface{}{
		"secret_id":  secret.ID,
		"project_id": projectID,
		"key":        key,
})
    if metrics.Default != nil { metrics.Default.ObserveSecrets("get", "success") }

	return secret, value, nil
}

// DeleteSecret removes a secret
func (s *SecretsService) DeleteSecret(ctx context.Context, tenantID, projectID, key, environment string) error {
	// Get secret first to verify tenant and get vault path
    secret, err := s.repo.GetByKey(ctx, projectID, key, environment)
    if err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("delete", "error") }
		return err
	}

	// Verify tenant isolation
    if secret.TenantID != tenantID {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("delete", "error") }
		return fmt.Errorf("tenant mismatch: unauthorized access")
	}

	// Delete from Vault
	vaultPath := secret.VaultPath()
    if err := s.vault.DeleteSecret(ctx, vaultPath); err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("delete", "error") }
		return fmt.Errorf("failed to delete from vault: %w", err)
	}

// Delete metadata from database
if err := s.repo.Delete(ctx, secret.ID); err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("delete", "error") }
		return fmt.Errorf("failed to delete metadata: %w", err)
}

// Audit & publish
if s.audit != nil { _ = s.audit.Write(ctx, &domain.AuditLogEntry{TenantID: secret.TenantID, ProjectID: projectID, Key: key, Environment: environment, Action: "deleted", Actor: "", OccurredAt: secret.UpdatedAt}) }
s.publishEvent("secret.deleted", map[string]interface{}{
		"secret_id":  secret.ID,
		"project_id": projectID,
		"key":        key,
})
    if metrics.Default != nil { metrics.Default.ObserveSecrets("delete", "success") }

	return nil
}

// ListSecrets returns paginated secrets (metadata only)
func (s *SecretsService) ListSecrets(ctx context.Context, projectID, environment string, limit int, cursor string) ([]*domain.Secret, string, error) {
    items, next, err := s.repo.List(ctx, projectID, environment, limit, cursor)
    if metrics.Default != nil {
        if err != nil { metrics.Default.ObserveSecrets("list", "error") } else { metrics.Default.ObserveSecrets("list", "success") }
    }
    return items, next, err
}

// UpdateSecret updates secret metadata (metadata-only example updates UpdatedAt)
func (s *SecretsService) UpdateSecret(ctx context.Context, secret *domain.Secret) error {
    if err := s.repo.Update(ctx, secret); err != nil {
        if metrics.Default != nil { metrics.Default.ObserveSecrets("update", "error") }
        return err
    }
	if s.audit != nil { _ = s.audit.Write(ctx, &domain.AuditLogEntry{TenantID: secret.TenantID, ProjectID: secret.ProjectID, Key: secret.Key, Environment: secret.Environment, Action: "updated", Actor: secret.CreatedBy, OccurredAt: secret.UpdatedAt}) }
	s.publishEvent("secret.updated", map[string]any{"secret_id": secret.ID, "project_id": secret.ProjectID, "key": secret.Key})
    if metrics.Default != nil { metrics.Default.ObserveSecrets("update", "success") }
	return nil
}

// publishEvent publishes an event (best effort)
func (s *SecretsService) publishEvent(eventType string, data map[string]interface{}) {
	if s.pub == nil { return }
	_ = s.pub.Publish(context.Background(), eventType, data)
}
