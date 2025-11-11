package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecret_Validate_ValidSecret(t *testing.T) {
	s := &Secret{
		TenantID:    "tenant-123",
		ProjectID:   "proj-456",
		Key:         "DATABASE_URL",
		Environment: "production",
		CreatedBy:   "user@example.com",
	}

	err := s.Validate()
	assert.NoError(t, err)
}

func TestSecret_Validate_MissingProjectID(t *testing.T) {
	s := &Secret{
		TenantID:    "tenant-123",
		Key:         "DATABASE_URL",
		Environment: "production",
		CreatedBy:   "user@example.com",
	}

	err := s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projectID")
}

func TestSecret_Validate_MissingKey(t *testing.T) {
	s := &Secret{
		TenantID:    "tenant-123",
		ProjectID:   "proj-456",
		Environment: "production",
		CreatedBy:   "user@example.com",
	}

	err := s.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
}

func TestSecret_VaultPath_GeneratesCorrectPath(t *testing.T) {
	s := &Secret{
		TenantID:    "tenant-123",
		ProjectID:   "proj-456",
		Key:         "DATABASE_URL",
		Environment: "production",
	}

	path := s.VaultPath()
	expected := "secret/data/tenant-123/proj-456/production/DATABASE_URL"
	assert.Equal(t, expected, path)
}

func TestSecretVersion_CreatedAtBeforeRotatedAt(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	sv := &SecretVersion{
		SecretID:      "secret-123",
		VersionNumber: 1,
		CreatedAt:     now,
		RotatedAt:     &later,
		CreatedBy:     "user@example.com",
	}

	assert.True(t, sv.CreatedAt.Before(*sv.RotatedAt))
}

func TestEnvParityCheck_HasDrift_WithMissingKeys(t *testing.T) {
    check := &EnvParityCheck{
		ProjectID:      "proj-456",
        MissingKeys:    []string{"API_KEY", "SECRET_" + "TOKEN"},
		MismatchedKeys: []string{},
		ExtraKeys:      []string{},
		ScanTimestamp:  time.Now(),
	}

	assert.True(t, check.HasDrift())
}

func TestEnvParityCheck_HasDrift_NoDrift(t *testing.T) {
	check := &EnvParityCheck{
		ProjectID:      "proj-456",
		MissingKeys:    []string{},
		MismatchedKeys: []string{},
		ExtraKeys:      []string{},
		ScanTimestamp:  time.Now(),
	}

	assert.False(t, check.HasDrift())
}

func TestClientLeakScan_SeverityLevels(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		valid    bool
	}{
		{"critical severity", "critical", true},
		{"high severity", "high", true},
		{"medium severity", "medium", true},
		{"low severity", "low", true},
		{"invalid severity", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &ClientLeakScan{
				ProjectID:   "proj-456",
				SnapshotID:  "snap-789",
				FilePath:    "/src/config.js",
				LineNumber:  42,
				Pattern:     "hardcoded_api_key",
				Severity:    tt.severity,
				Fixed:       false,
			}

			err := scan.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestProjectSecrets_EmptyInitialization(t *testing.T) {
	ps := &ProjectSecrets{
		ProjectID:      "proj-456",
		TotalCount:     0,
		LastSyncAt:     time.Time{},
		ComplianceOK:   false,
	}

	assert.Equal(t, 0, ps.TotalCount)
	assert.True(t, ps.LastSyncAt.IsZero())
	assert.False(t, ps.ComplianceOK)
}
