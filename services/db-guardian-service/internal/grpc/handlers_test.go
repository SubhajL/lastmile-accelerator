package grpcserver

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/dbguardian/v1"
	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/dto"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/server"
	"example.com/lma/db-guardian-service/internal/service"
)

type fakeConnSvc struct{
	regCalled bool
	lastConn *models.DBConnection
	lastMakeDefault bool
	id string
	list []models.DBConnection
	err error
}
func (f *fakeConnSvc) RegisterConnection(ctx context.Context, c *models.DBConnection, makeDefault bool) (string, error) { f.regCalled = true; f.lastConn = c; f.lastMakeDefault = makeDefault; return f.id, f.err }
func (f *fakeConnSvc) DeleteConnection(ctx context.Context, id string) error { return nil }
func (f *fakeConnSvc) GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error) { return nil, nil }
func (f *fakeConnSvc) ListConnections(ctx context.Context, projectID string) ([]models.DBConnection, error) { return f.list, f.err }

type fakeMigGuard struct{ res *analyzer.MigrationValidationResult; err error }
func (f *fakeMigGuard) ValidateMigration(ctx context.Context, sql string, opts analyzer.ValidationOptions) (*analyzer.MigrationValidationResult, error) { return f.res, f.err }

type fakeAnalysis struct{ rep *service.AnalysisReport; err error }
func (f *fakeAnalysis) RunFullAnalysis(ctx context.Context, projectID, migrationName, migrationSQL string, role analyzer.AnalyzeOptions, val analyzer.ValidationOptions, idx analyzer.IndexAnalysisOptions) (*service.AnalysisReport, error) { return f.rep, f.err }

type fakeRecs struct{ idx []analyzer.IndexRecommendation; role *analyzer.RoleAnalysisResult; err error }
func (f *fakeRecs) Get(ctx context.Context, projectID string, onlyUnapplied bool) ([]analyzer.IndexRecommendation, *analyzer.RoleAnalysisResult, error) { return f.idx, f.role, f.err }

type fakePolicy struct{ p *models.RolePolicy; err error }
func (f *fakePolicy) GetPolicy(ctx context.Context, projectID string) (*models.RolePolicy, error) { return f.p, f.err }
func (f *fakePolicy) UpdatePolicy(ctx context.Context, projectID, yaml string) (*models.RolePolicy, error) { return f.p, f.err }

type fakeDrift struct{ d *dto.DriftResponse; err error }
func (f *fakeDrift) Check(ctx context.Context, projectID string) (*dto.DriftResponse, error) { return f.d, f.err }

func TestRegisterConnection_Delegates_Service(t *testing.T) {
	deps := &server.Dependencies{}
	fc := &fakeConnSvc{id: "123"}
	deps.ConnSvc = fc
	g := newGuardianServer(deps)
	resp, err := g.RegisterConnection(context.Background(), &dbguardianv1.RegisterConnectionRequest{ProjectId: "p1", Name: "n1", Driver: "postgres", DsnRef: "v:ref", MakeDefault: true})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if resp.GetId() != "123" { t.Fatalf("want id 123, got %s", resp.GetId()) }
	if !fc.regCalled || fc.lastConn == nil || !fc.lastMakeDefault { t.Fatalf("service not called properly: %#v", fc) }
}

func TestListConnections_MapsModels(t *testing.T) {
	deps := &server.Dependencies{}
	deps.ConnSvc = &fakeConnSvc{list: []models.DBConnection{{ID: "1", ProjectID: "p", Name: "n", Driver: "postgres", DSNRef: "r", IsDefault: true}}}
	g := newGuardianServer(deps)
	resp, err := g.ListConnections(context.Background(), &dbguardianv1.ListConnectionsRequest{ProjectId: "p"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if len(resp.GetConnections()) != 1 { t.Fatalf("want 1, got %d", len(resp.GetConnections())) }
	c := resp.GetConnections()[0]
	if c.GetId() != "1" || c.GetProjectId() != "p" || !c.GetIsDefault() { t.Fatalf("bad mapping: %#v", c) }
}

func TestValidateMigration_ReturnsStatusAndFindingsJSON(t *testing.T) {
	deps := &server.Dependencies{}
	deps.MigGuard = &fakeMigGuard{res: &analyzer.MigrationValidationResult{Status: "warn", Findings: []analyzer.Finding{{Category:"performance"}}, ValidatedAt: time.Now()}}
	g := newGuardianServer(deps)
	resp, err := g.ValidateMigration(context.Background(), &dbguardianv1.ValidateMigrationRequest{ProjectId: "p", MigrationName:"m1", Sql:"SQL", CheckBreaking:true, CheckPerformance:true, MaxTableSizeBytes: 1024})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if resp.GetStatus() != "warn" { t.Fatalf("want warn, got %s", resp.GetStatus()) }
	var findings []analyzer.Finding
	if err := json.Unmarshal([]byte(resp.GetFindingsJson()), &findings); err != nil { t.Fatalf("json: %v", err) }
	if len(findings) != 1 || findings[0].Category != "performance" { t.Fatalf("bad findings json: %s", resp.GetFindingsJson()) }
}

func TestRunAnalysis_EncodesReportJSON(t *testing.T) {
	deps := &server.Dependencies{}
	deps.AnalysisSvc = &fakeAnalysis{rep: &service.AnalysisReport{Role: &analyzer.RoleAnalysisResult{RolesAnalyzed: 2}, Migration: &analyzer.MigrationValidationResult{Status:"pass"}, Index: &analyzer.IndexRecommendations{Recommendations: []analyzer.IndexRecommendation{{TableName:"t", Columns: []string{"a"}}}}}}
	g := newGuardianServer(deps)
	resp, err := g.RunAnalysis(context.Background(), &dbguardianv1.RunAnalysisRequest{ProjectId: "p", MigrationName: "m", Sql: "SQL"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	var out map[string]any
	if err := json.Unmarshal([]byte(resp.GetReportJson()), &out); err != nil { t.Fatalf("json: %v", err) }
	if out["role"] == nil || out["migration"] == nil || out["index"] == nil { t.Fatalf("missing keys: %v", out) }
}

func TestGetRecommendations_Merges(t *testing.T) {
	deps := &server.Dependencies{}
	deps.RecsProvider = &fakeRecs{idx: []analyzer.IndexRecommendation{{TableName:"t", Columns: []string{"c"}, BenefitScore: 5, Reason: "foo"}}, role: &analyzer.RoleAnalysisResult{RolesAnalyzed: 3}}
	g := newGuardianServer(deps)
	resp, err := g.GetRecommendations(context.Background(), &dbguardianv1.GetRecommendationsRequest{ProjectId: "p", OnlyUnapplied: true})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if len(resp.GetIndex()) != 1 || resp.GetRole() == nil || resp.GetRole().GetRolesAnalyzed() != 3 { t.Fatalf("bad: %#v", resp) }
}

func TestGetPolicy_And_UpdatePolicy(t *testing.T) {
	deps := &server.Dependencies{}
	deps.PolicyMgr = &fakePolicy{p: &models.RolePolicy{ProjectID: "p", SpecYAML: "a: b", Version: 2}}
	g := newGuardianServer(deps)
	gp, err := g.GetPolicy(context.Background(), &dbguardianv1.GetPolicyRequest{ProjectId: "p"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	up, err := g.UpdatePolicy(context.Background(), &dbguardianv1.UpdatePolicyRequest{ProjectId: "p", SpecYaml: "a: b"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if !reflect.DeepEqual(gp, &dbguardianv1.GetPolicyResponse{ProjectId: "p", SpecYaml: "a: b", Version: 2}) { t.Fatalf("bad get: %#v", gp) }
	if !reflect.DeepEqual(up, &dbguardianv1.UpdatePolicyResponse{ProjectId: "p", SpecYaml: "a: b", Version: 2}) { t.Fatalf("bad update: %#v", up) }
}

func TestCheckDrift_MissingAndExtra(t *testing.T) {
	deps := &server.Dependencies{}
	deps.DriftCheck = &fakeDrift{d: &dto.DriftResponse{MissingIndexes: []string{"i1"}, ExtraIndexes: []string{"i2"}}}
	g := newGuardianServer(deps)
	resp, err := g.CheckDrift(context.Background(), &dbguardianv1.CheckDriftRequest{ProjectId: "p"})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if len(resp.GetMissingIndexes()) != 1 || len(resp.GetExtraIndexes()) != 1 { t.Fatalf("bad: %#v", resp) }
}