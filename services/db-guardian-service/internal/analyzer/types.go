package analyzer

import "time"

// Severity levels for findings
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarn     Severity = "warn"
	SeverityInfo     Severity = "info"
)

// TableInfo represents metadata about a database table
type TableInfo struct {
	Schema    string
	Name      string
	RowCount  int64
	SizeBytes int64
}

// ColumnInfo represents metadata about a table column
type ColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue string
	IsPrimaryKey bool
}

// IndexInfo represents metadata about a database index
type IndexInfo struct {
	Name      string
	TableName string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
}

// RoleInfo represents a database role with permissions
type RoleInfo struct {
	Name         string
	IsSuperuser  bool
	CanLogin     bool
	Permissions  []Permission
}

// Permission represents a granted permission
type Permission struct {
	Schema     string
	ObjectType string // table, schema, database
	ObjectName string
	Privilege  string // SELECT, INSERT, UPDATE, DELETE, etc.
}

// QueryStat represents query execution statistics
type QueryStat struct {
	Query          string
	Calls          int64
	TotalTime      float64
	MeanTime       float64
	TablesAccessed []string
}

// MigrationAnalysis represents analysis result of migration SQL
type MigrationAnalysis struct {
	SQL             string
	Operations      []SQLOperation
	AffectedTables  []string
	HasBreaking     bool
	EstimatedImpact string
}

// SQLOperation represents a parsed SQL operation
type SQLOperation struct {
	Type       string // ALTER, DROP, CREATE, etc.
	ObjectType string // TABLE, COLUMN, INDEX, etc.
	ObjectName string
	Details    map[string]string
}

// Finding represents an analysis finding
type Finding struct {
	Severity    Severity
	Category    string
	Title       string
	Description string
	Remediation string
}

// RoleFinding extends Finding with role-specific information
type RoleFinding struct {
	Finding
	RoleName string
}

// AnalyzeOptions contains options for role analysis
type AnalyzeOptions struct {
	IncludeSystemRoles bool
	CheckPublicSchema  bool
}

// RoleAnalysisResult contains role analysis results
type RoleAnalysisResult struct {
	Findings       []RoleFinding
	PolicyYAML     string
	AnalyzedAt     time.Time
	RolesAnalyzed  int
}

// ValidationOptions contains options for migration validation
type ValidationOptions struct {
	CheckBreaking    bool
	CheckPerformance bool
	MaxTableSize     int64 // bytes, for lock warnings
}

// MigrationValidationResult contains migration validation results
type MigrationValidationResult struct {
	Status          string // pass, warn, fail
	Findings        []Finding
	Rollback        *RollbackAssessment
	ValidatedAt     time.Time
}

// RollbackAssessment evaluates rollback safety
type RollbackAssessment struct {
	IsReversible bool
	RollbackSQL  string
	Notes        string
}

// IndexAnalysisOptions contains options for index analysis
type IndexAnalysisOptions struct {
	MinQueryExecutions int
	MinTableSize       int64
	CheckDuplicates    bool
}

// IndexRecommendations contains index analysis results
type IndexRecommendations struct {
	Recommendations []IndexRecommendation
	Duplicates      []DuplicateFinding
	AnalyzedAt      time.Time
}

// IndexRecommendation represents a suggested index
type IndexRecommendation struct {
	TableName    string
	Columns      []string
	Reason       string
	BenefitScore int
	EstimatedImpact string
}

// DuplicateFinding represents redundant indexes
type DuplicateFinding struct {
	RedundantIndex string
	CoveredBy      string
	Reason         string
}
