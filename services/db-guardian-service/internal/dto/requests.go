package dto

import (
	"fmt"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/models"
)

// RegisterConnectionRequest represents API input to create/register a DB connection.
type RegisterConnectionRequest struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Driver      string `json:"driver"`
	DSNRef      string `json:"dsn_ref"`
	MakeDefault bool   `json:"make_default"`
}

func (r RegisterConnectionRequest) Validate() error {
	if r.ProjectID == "" || r.Name == "" || r.Driver == "" || r.DSNRef == "" {
		return fmt.Errorf("project_id, name, driver, dsn_ref are required")
	}
	// Accept only supported driver(s) for now
	switch r.Driver {
	case "postgres":
		return nil
	default:
		return fmt.Errorf("unsupported driver: %s", r.Driver)
	}
}

func (r RegisterConnectionRequest) ToModel() models.DBConnection {
	return models.DBConnection{
		ProjectID: r.ProjectID,
		Name:      r.Name,
		Driver:    r.Driver,
		DSNRef:    r.DSNRef,
		IsDefault: r.MakeDefault,
	}
}

// ValidateMigrationRequest represents a request to validate a migration.
type ValidateMigrationRequest struct {
	ProjectID        string `json:"project_id"`
	MigrationName    string `json:"migration_name"`
	SQL              string `json:"sql"`
	CheckBreaking    bool   `json:"check_breaking"`
	CheckPerformance bool   `json:"check_performance"`
	MaxTableSizeBytes int64 `json:"max_table_size_bytes"`
}

func (r ValidateMigrationRequest) Validate() error {
	if r.ProjectID == "" || r.MigrationName == "" {
		return fmt.Errorf("project_id and migration_name are required")
	}
	return nil
}

func (r ValidateMigrationRequest) ToValidationOptions(cfg *config.Config) analyzer.ValidationOptions {
	maxSize := r.MaxTableSizeBytes
	if maxSize <= 0 && cfg != nil {
		maxSize = cfg.AnalysisMaxLockWarnTableSizeBytes
	}
	return analyzer.ValidationOptions{
		CheckBreaking:    r.CheckBreaking,
		CheckPerformance: r.CheckPerformance,
		MaxTableSize:     maxSize,
	}
}

// RunAnalysisRequest requests a full analysis run.
type RunAnalysisRequest struct {
	ProjectID     string                         `json:"project_id"`
	MigrationName string                         `json:"migration_name"`
	SQL           string                         `json:"sql"`
	Role          analyzer.AnalyzeOptions        `json:"role_options"`
	Validation    ValidateMigrationRequest       `json:"validation_options"`
	Index         analyzer.IndexAnalysisOptions  `json:"index_options"`
}

func (r RunAnalysisRequest) Validate() error {
	if r.ProjectID == "" || r.MigrationName == "" {
		return fmt.Errorf("project_id and migration_name are required")
	}
	return nil
}

func (r RunAnalysisRequest) ToOptions(cfg *config.Config) (analyzer.AnalyzeOptions, analyzer.ValidationOptions, analyzer.IndexAnalysisOptions) {
	role := r.Role
	val := r.Validation.ToValidationOptions(cfg)
	idx := r.Index
	return role, val, idx
}
