package dto

import (
	"encoding/json"
	"testing"

	"example.com/lma/db-guardian-service/internal/analyzer"
)

func TestResponse_Mapping_ConnectionResponse(t *testing.T) {
	in := analyzer.RoleAnalysisResult{} // not used here; just avoid unused import issues
	_ = in

	rec := ConnectionResponse{
		ID:        "c1",
		ProjectID: "p1",
		Name:      "primary",
		Driver:    "postgres",
		DSNRef:    "vault:path",
		IsDefault: true,
	}
	b, err := json.Marshal(rec)
	if err != nil || len(b) == 0 {
		t.Fatalf("marshal error: %v", err)
	}
}

func TestNewMigrationValidationResponse_MapsFields(t *testing.T) {
	in := &analyzer.MigrationValidationResult{
		Status:   "fail",
		Findings: []analyzer.Finding{{Severity: analyzer.SeverityCritical, Category: "breaking", Title: "drop"}},
	}
	out := NewMigrationValidationResponse(in)
	if out.Status != "fail" || len(out.Findings) != 1 {
		t.Fatalf("mapping incorrect")
	}
}

func TestNewIndexRecommendationsResponse_MapsRecommendations(t *testing.T) {
	in := &analyzer.IndexRecommendations{
		Recommendations: []analyzer.IndexRecommendation{{TableName: "users", Columns: []string{"email"}, BenefitScore: 10}},
	}
	out := NewIndexRecommendationsResponse(in)
	if len(out.Recommendations) != 1 || out.Recommendations[0].TableName != "users" {
		t.Fatalf("mapping incorrect")
	}
}

func TestNewAnalysisReportResponse_CombinesAll(t *testing.T) {
	role := &analyzer.RoleAnalysisResult{}
	mig := &analyzer.MigrationValidationResult{Status: "pass"}
	idx := &analyzer.IndexRecommendations{}
	res := NewAnalysisReportResponse(role, mig, idx)
	if res.Migration.Status != "pass" {
		t.Fatalf("expected pass status in response")
	}
}
