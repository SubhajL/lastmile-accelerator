package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/service"
)

type fakeConnSvc struct{}

func (f *fakeConnSvc) RegisterConnection(ctx context.Context, conn *models.DBConnection, makeDefault bool) (string, error) {
	return "conn-xyz", nil
}
func (f *fakeConnSvc) DeleteConnection(ctx context.Context, id string) error { return nil }
func (f *fakeConnSvc) GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error) {
	return &models.DBConnection{ID: "conn-xyz", ProjectID: projectID, Name: "primary", Driver: "postgres", DSNRef: "vault:path", IsDefault: true}, nil
}
func (f *fakeConnSvc) ListConnections(ctx context.Context, projectID string) ([]models.DBConnection, error) {
	return []models.DBConnection{{ID: "conn-xyz", ProjectID: projectID, Name: "primary", Driver: "postgres", DSNRef: "vault:path", IsDefault: true}}, nil
}

type fakeAnalysisSvc struct{}

func (f *fakeAnalysisSvc) RunFullAnalysis(ctx context.Context, projectID, migrationName, migrationSQL string, role analyzer.AnalyzeOptions, val analyzer.ValidationOptions, idx analyzer.IndexAnalysisOptions) (*service.AnalysisReport, error) {
	return &service.AnalysisReport{
		Role:      &analyzer.RoleAnalysisResult{RolesAnalyzed: 1},
		Migration: &analyzer.MigrationValidationResult{Status: "pass"},
		Index:     &analyzer.IndexRecommendations{Recommendations: []analyzer.IndexRecommendation{{TableName: "users", Columns: []string{"email"}}}},
	}, nil
}

type fakeMigGuard struct{}

func (f *fakeMigGuard) ValidateMigration(ctx context.Context, sql string, opts analyzer.ValidationOptions) (*analyzer.MigrationValidationResult, error) {
	return &analyzer.MigrationValidationResult{Status: "pass"}, nil
}

type fakeRecs struct{}

func (f *fakeRecs) Get(ctx context.Context, projectID string, onlyUnapplied bool) ([]analyzer.IndexRecommendation, *analyzer.RoleAnalysisResult, error) {
	return []analyzer.IndexRecommendation{{TableName: "users", Columns: []string{"email"}}}, &analyzer.RoleAnalysisResult{RolesAnalyzed: 1}, nil
}

func TestPOST_RegisterConnection_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
	deps := &Dependencies{ConnSvc: &fakeConnSvc{}}
	srv := New(cfg, deps)

	reqBody := map[string]any{
		"project_id": "p1",
		"name": "primary",
		"driver": "postgres",
		"dsn_ref": "vault:path",
		"make_default": true,
	}
	b, _ := json.Marshal(reqBody)
r := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/db/connections", bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("conn-xyz")) {
		t.Fatalf("expected response to contain id")
	}
}

func TestDELETE_Connection_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
	deps := &Dependencies{ConnSvc: &fakeConnSvc{}}
	srv := New(cfg, deps)

	r := httptest.NewRequest(http.MethodDelete, "/api/connections/conn-xyz", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestGET_DefaultConnection_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
	deps := &Dependencies{ConnSvc: &fakeConnSvc{}}
	srv := New(cfg, deps)

	r := httptest.NewRequest(http.MethodGet, "/api/connections/default?project_id=p1", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("\"project_id\":\"p1\"")) {
		t.Fatalf("expected project_id in response")
	}
}

func TestPOST_ValidateMigration_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
	deps := &Dependencies{MigGuard: &fakeMigGuard{}}
	srv := New(cfg, deps)

	reqBody := map[string]any{"project_id": "p1", "migration_name": "001", "sql": "CREATE TABLE x(id int)", "check_breaking": true}
	b, _ := json.Marshal(reqBody)
r := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/db/migrations/validate", bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("\"status\":\"pass\"")) {
		t.Fatalf("expected pass status in response")
	}
}

func TestGET_Recommendations_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
// fake recs provider
	deps := &Dependencies{RecsProvider: &fakeRecs{}}
	srv := New(cfg, deps)

	r := httptest.NewRequest(http.MethodGet, "/v1/projects/p1/db/recommendations", nil)
	w := httptest.NewRecorder()
	
	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPOST_RunAnalysis_Success(t *testing.T) {
	cfg := &config.Config{ServicePort: "0", ServiceName: "db-guardian-service"}
	deps := &Dependencies{AnalysisSvc: &fakeAnalysisSvc{}}
	srv := New(cfg, deps)

	reqBody := map[string]any{"project_id": "p1", "migration_name": "002", "sql": "ALTER TABLE x ADD y int"}
	b, _ := json.Marshal(reqBody)
r := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/db/analyze", bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("\"migration\":{")) || !bytes.Contains(w.Body.Bytes(), []byte("\"status\":\"pass\"")) {
		t.Fatalf("expected migration status in full report response")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("\"index\":{")) || !bytes.Contains(w.Body.Bytes(), []byte("users")) {
		t.Fatalf("expected index recommendations in full report response")
	}
}
