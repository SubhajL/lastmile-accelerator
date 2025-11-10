package domain

import (
	"fmt"
	"time"
)

// Secret represents secret metadata stored in the database
type Secret struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ProjectID   string    `json:"project_id"`
	Key         string    `json:"key"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
}

// Validate checks if all required fields are present
func (s *Secret) Validate() error {
	if s.TenantID == "" {
		return fmt.Errorf("tenantID is required")
	}
	if s.ProjectID == "" {
		return fmt.Errorf("projectID is required")
	}
	if s.Key == "" {
		return fmt.Errorf("key is required")
	}
	if s.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if s.CreatedBy == "" {
		return fmt.Errorf("createdBy is required")
	}
	return nil
}

// VaultPath constructs the Vault path for this secret
func (s *Secret) VaultPath() string {
	return fmt.Sprintf("secret/data/%s/%s/%s/%s", s.TenantID, s.ProjectID, s.Environment, s.Key)
}

// SecretVersion tracks version history for secrets
type SecretVersion struct {
	ID            string     `json:"id"`
	SecretID      string     `json:"secret_id"`
	VersionNumber int        `json:"version_number"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedBy     string     `json:"created_by"`
	RotatedAt     *time.Time `json:"rotated_at,omitempty"`
}

// EnvironmentConfig defines expected secrets for an environment
type EnvironmentConfig struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Environment string    `json:"environment"`
	ExpectedKeys []string  `json:"expected_keys"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvParityCheck records drift detection results
type EnvParityCheck struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	ScanTimestamp  time.Time `json:"scan_timestamp"`
	MissingKeys    []string  `json:"missing_keys"`
	MismatchedKeys []string  `json:"mismatched_keys"`
	ExtraKeys      []string  `json:"extra_keys"`
}

// HasDrift returns true if any drift was detected
func (e *EnvParityCheck) HasDrift() bool {
	return len(e.MissingKeys) > 0 || len(e.MismatchedKeys) > 0 || len(e.ExtraKeys) > 0
}

// ClientLeakScan stores scan results for hardcoded secrets
type ClientLeakScan struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	SnapshotID string    `json:"snapshot_id"`
	FilePath   string    `json:"file_path"`
	LineNumber int       `json:"line_number"`
	Pattern    string    `json:"pattern"`
	Severity   string    `json:"severity"`
	Fixed      bool      `json:"fixed"`
	CreatedAt  time.Time `json:"created_at"`
}

// ValidSeverities defines allowed severity levels
var ValidSeverities = map[string]bool{
	"critical": true,
	"high":     true,
	"medium":   true,
	"low":      true,
}

// Validate checks if the scan has valid fields
func (c *ClientLeakScan) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("projectID is required")
	}
	if c.SnapshotID == "" {
		return fmt.Errorf("snapshotID is required")
	}
	if c.FilePath == "" {
		return fmt.Errorf("filePath is required")
	}
	if !ValidSeverities[c.Severity] {
		return fmt.Errorf("invalid severity: %s (must be critical, high, medium, or low)", c.Severity)
	}
	return nil
}

// ProjectSecrets aggregates project-level secret metadata
type ProjectSecrets struct {
	ProjectID    string    `json:"project_id"`
	TotalCount   int       `json:"total_count"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	ComplianceOK bool      `json:"compliance_ok"`
}
