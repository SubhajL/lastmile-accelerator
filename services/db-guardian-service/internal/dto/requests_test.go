package dto

import (
	"encoding/json"
	"testing"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/config"
)

func TestRegisterConnectionRequest_Validate_Success(t *testing.T) {
	req := RegisterConnectionRequest{
		ProjectID:  "p1",
		Name:       "primary",
		Driver:     "postgres",
		DSNRef:     "secret/data/dbs/p1/primary",
		MakeDefault: true,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterConnectionRequest_Validate_MissingFields_Error(t *testing.T) {
	tests := []RegisterConnectionRequest{
		{ProjectID: "", Name: "n", Driver: "postgres", DSNRef: "x"},
		{ProjectID: "p", Name: "", Driver: "postgres", DSNRef: "x"},
		{ProjectID: "p", Name: "n", Driver: "", DSNRef: "x"},
		{ProjectID: "p", Name: "n", Driver: "postgres", DSNRef: ""},
		{ProjectID: "p", Name: "n", Driver: "mysql", DSNRef: "x"},
	}
	for i, req := range tests {
		if err := req.Validate(); err == nil {
			t.Fatalf("case %d: expected error, got nil", i)
		}
	}
}

func TestRegisterConnectionRequest_ToModel_SetsIsDefault(t *testing.T) {
	req := RegisterConnectionRequest{
		ProjectID:  "p1",
		Name:       "primary",
		Driver:     "postgres",
		DSNRef:     "secret/data/dbs/p1/primary",
		MakeDefault: true,
	}
	m := req.ToModel()
	if !m.IsDefault {
		t.Fatalf("expected IsDefault true, got false")
	}
}

func TestValidateMigrationRequest_ToOptions_UsesOverridesOrDefaults(t *testing.T) {
	cfg := &config.Config{AnalysisMaxLockWarnTableSizeBytes: 64 * 1024 * 1024}
	req := ValidateMigrationRequest{
		ProjectID:   "p1",
		MigrationName: "001",
		SQL:           "ALTER TABLE users ADD COLUMN x int",
		CheckBreaking: true,
		CheckPerformance: true,
		MaxTableSizeBytes: 0, // use default
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	opts := req.ToValidationOptions(cfg)
	if !opts.CheckBreaking || !opts.CheckPerformance {
		t.Fatalf("expected checks enabled")
	}
	if opts.MaxTableSize != cfg.AnalysisMaxLockWarnTableSizeBytes {
		t.Fatalf("expected default MaxTableSize from config")
	}
}

func TestRunAnalysisRequest_Validate_RequiresProjectAndMigrationName(t *testing.T) {
	req := RunAnalysisRequest{}
	if err := req.Validate(); err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestRunAnalysisRequest_ToOptions_ReturnsExpectedTriples(t *testing.T) {
	cfg := &config.Config{AnalysisMaxLockWarnTableSizeBytes: 50 * 1024 * 1024}
	req := RunAnalysisRequest{
		ProjectID:     "p1",
		MigrationName: "002",
		SQL:           "ALTER TABLE users DROP COLUMN email",
		Role:          analyzer.AnalyzeOptions{IncludeSystemRoles: true, CheckPublicSchema: true},
		Validation:    ValidateMigrationRequest{CheckBreaking: true, CheckPerformance: true},
		Index:         analyzer.IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1, CheckDuplicates: true},
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	role, val, idx := req.ToOptions(cfg)
	if !role.IncludeSystemRoles || !role.CheckPublicSchema {
		t.Fatalf("role options not mapped correctly")
	}
	if !val.CheckBreaking || !val.CheckPerformance || val.MaxTableSize == 0 {
		t.Fatalf("validation options not mapped correctly")
	}
	if idx.MinQueryExecutions != 100 || !idx.CheckDuplicates {
		t.Fatalf("index options not mapped correctly")
	}
}

func TestRequests_JSONTags_AreCorrect(t *testing.T) {
	req := RegisterConnectionRequest{ProjectID: "p", Name: "n", Driver: "postgres", DSNRef: "r", MakeDefault: true}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	// quick smoke check for tag names
	js := string(b)
	for _, k := range []string{"project_id","name","driver","dsn_ref","make_default"} {
		if !contains(js, `"`+k+`"`) {
			t.Fatalf("expected json key %s in output", k)
		}
	}
}

// tiny helper to avoid importing strings
func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || s[:len(sub)] == sub))) }
