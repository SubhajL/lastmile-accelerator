package models

import "time"

// DBConnection represents a database connection reference for a project
type DBConnection struct {
	ID         string
	ProjectID  string
	Name       string
	Driver     string // postgres, mysql, sqlite
	DSNRef     string // Vault path or reference, never plaintext credentials
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// RolePolicy represents a role-based access control policy for a project
type RolePolicy struct {
	ID        string
	ProjectID string
	SpecYAML  string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MigrationAudit represents a record of migration validation
type MigrationAudit struct {
	ID            string
	ProjectID     string
	MigrationName string
	Status        string // pass, warn, fail
	FindingsJSON  string // JSON-encoded findings
	CreatedAt     time.Time
}

// IndexRecommendation represents a suggested database index
type IndexRecommendation struct {
	ID           string
	ProjectID    string
	TableName    string
	ColumnNames  []string
	Reason       string
	BenefitScore int
	Applied      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
