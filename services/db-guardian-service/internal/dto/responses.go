package dto

import (
	"example.com/lma/db-guardian-service/internal/analyzer"
)

// ConnectionResponse represents a DB connection for API responses.
type ConnectionResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Driver    string `json:"driver"`
	DSNRef    string `json:"dsn_ref"`
	IsDefault bool   `json:"is_default"`
}

// MigrationValidationResponse wraps analyzer.MigrationValidationResult for API shape.
type MigrationValidationResponse struct {
	Status   string            `json:"status"`
	Findings []analyzer.Finding `json:"findings"`
	// Rollback and timestamps can be added as needed
	Rollback *analyzer.RollbackAssessment `json:"rollback,omitempty"`
}

// IndexRecommendationsResponse wraps index recommendations
 type IndexRecommendationsResponse struct {
	Recommendations []analyzer.IndexRecommendation `json:"recommendations"`
	Duplicates      []analyzer.DuplicateFinding    `json:"duplicates,omitempty"`
}

// AnalysisReportResponse combines all analyzer results for a full analysis.
type AnalysisReportResponse struct {
	Role      *analyzer.RoleAnalysisResult         `json:"role"`
	Migration *MigrationValidationResponse        `json:"migration"`
	Index     *IndexRecommendationsResponse       `json:"index"`
}

func NewMigrationValidationResponse(in *analyzer.MigrationValidationResult) *MigrationValidationResponse {
	if in == nil { return nil }
	return &MigrationValidationResponse{Status: in.Status, Findings: in.Findings, Rollback: in.Rollback}
}

func NewIndexRecommendationsResponse(in *analyzer.IndexRecommendations) *IndexRecommendationsResponse {
	if in == nil { return nil }
	return &IndexRecommendationsResponse{Recommendations: in.Recommendations, Duplicates: in.Duplicates}
}

func NewAnalysisReportResponse(role *analyzer.RoleAnalysisResult, mig *analyzer.MigrationValidationResult, idx *analyzer.IndexRecommendations) *AnalysisReportResponse {
	return &AnalysisReportResponse{
		Role:      role,
		Migration: NewMigrationValidationResponse(mig),
		Index:     NewIndexRecommendationsResponse(idx),
	}
}

// PolicyResponse represents a stored role policy
 type PolicyResponse struct {
	ProjectID string `json:"project_id"`
	SpecYAML  string `json:"spec_yaml"`
	Version   int    `json:"version"`
}

// DriftResponse represents schema/role drift findings
 type DriftResponse struct {
	MissingIndexes []string `json:"missing_indexes"`
	ExtraIndexes   []string `json:"extra_indexes"`
	RoleDrift      []string `json:"role_drift"`
}
