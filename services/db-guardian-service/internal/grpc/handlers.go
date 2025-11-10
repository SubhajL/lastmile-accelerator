package grpcserver

import (
	"context"
	"encoding/json"

	"example.com/lma/db-guardian-service/dbguardian/v1"
	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GuardianGRPCServer struct {
	dbguardianv1.UnimplementedDbGuardianServiceServer
	deps *server.Dependencies
}

func newGuardianServer(deps *server.Dependencies) dbguardianv1.DbGuardianServiceServer {
	return &GuardianGRPCServer{deps: deps}
}

func (g *GuardianGRPCServer) RegisterConnection(ctx context.Context, req *dbguardianv1.RegisterConnectionRequest) (*dbguardianv1.RegisterConnectionResponse, error) {
	if g.deps == nil || g.deps.ConnSvc == nil { return nil, status.Error(codes.Unavailable, "connection service unavailable") }
	m := &models.DBConnection{ProjectID: req.GetProjectId(), Name: req.GetName(), Driver: req.GetDriver(), DSNRef: req.GetDsnRef(), IsDefault: req.GetMakeDefault()}
	id, err := g.deps.ConnSvc.RegisterConnection(ctx, m, req.GetMakeDefault())
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	return &dbguardianv1.RegisterConnectionResponse{Id: id}, nil
}

func (g *GuardianGRPCServer) ListConnections(ctx context.Context, req *dbguardianv1.ListConnectionsRequest) (*dbguardianv1.ListConnectionsResponse, error) {
	if g.deps == nil || g.deps.ConnSvc == nil { return nil, status.Error(codes.Unavailable, "connection service unavailable") }
	list, err := g.deps.ConnSvc.ListConnections(ctx, req.GetProjectId())
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	out := make([]*dbguardianv1.Connection, 0, len(list))
	for _, c := range list {
		out = append(out, &dbguardianv1.Connection{Id: c.ID, ProjectId: c.ProjectID, Name: c.Name, Driver: c.Driver, DsnRef: c.DSNRef, IsDefault: c.IsDefault})
	}
	return &dbguardianv1.ListConnectionsResponse{Connections: out}, nil
}

func (g *GuardianGRPCServer) ValidateMigration(ctx context.Context, req *dbguardianv1.ValidateMigrationRequest) (*dbguardianv1.ValidateMigrationResponse, error) {
	if g.deps == nil || g.deps.MigGuard == nil { return nil, status.Error(codes.Unavailable, "migration validator unavailable") }
	opts := analyzer.ValidationOptions{CheckBreaking: req.GetCheckBreaking(), CheckPerformance: req.GetCheckPerformance(), MaxTableSize: req.GetMaxTableSizeBytes()}
	res, err := g.deps.MigGuard.ValidateMigration(ctx, req.GetSql(), opts)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	findings, _ := json.Marshal(res.Findings)
	return &dbguardianv1.ValidateMigrationResponse{Status: res.Status, FindingsJson: string(findings)}, nil
}

func (g *GuardianGRPCServer) RunAnalysis(ctx context.Context, req *dbguardianv1.RunAnalysisRequest) (*dbguardianv1.RunAnalysisResponse, error) {
	if g.deps == nil || g.deps.AnalysisSvc == nil { return nil, status.Error(codes.Unavailable, "analysis service unavailable") }
	role := analyzer.AnalyzeOptions{}
	val := analyzer.ValidationOptions{}
	idx := analyzer.IndexAnalysisOptions{}
	rep, err := g.deps.AnalysisSvc.RunFullAnalysis(ctx, req.GetProjectId(), req.GetMigrationName(), req.GetSql(), role, val, idx)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	wire := map[string]any{
		"role": rep.Role,
		"migration": rep.Migration,
		"index": rep.Index,
	}
	data, _ := json.Marshal(wire)
	return &dbguardianv1.RunAnalysisResponse{ReportJson: string(data)}, nil
}

func (g *GuardianGRPCServer) GetRecommendations(ctx context.Context, req *dbguardianv1.GetRecommendationsRequest) (*dbguardianv1.GetRecommendationsResponse, error) {
	if g.deps == nil || g.deps.RecsProvider == nil { return nil, status.Error(codes.Unavailable, "recommendations unavailable") }
	idx, role, err := g.deps.RecsProvider.Get(ctx, req.GetProjectId(), req.GetOnlyUnapplied())
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	var outIdx []*dbguardianv1.IndexRecommendation
	for _, r := range idx {
		outIdx = append(outIdx, &dbguardianv1.IndexRecommendation{TableName: r.TableName, Columns: r.Columns, BenefitScore: int32(r.BenefitScore), Reason: r.Reason})
	}
	var roleRec *dbguardianv1.RoleRecommendation
	if role != nil { roleRec = &dbguardianv1.RoleRecommendation{RolesAnalyzed: int32(role.RolesAnalyzed)} }
	return &dbguardianv1.GetRecommendationsResponse{Index: outIdx, Role: roleRec}, nil
}

func (g *GuardianGRPCServer) GetPolicy(ctx context.Context, req *dbguardianv1.GetPolicyRequest) (*dbguardianv1.GetPolicyResponse, error) {
	if g.deps == nil || g.deps.PolicyMgr == nil { return nil, status.Error(codes.Unavailable, "policy manager unavailable") }
	p, err := g.deps.PolicyMgr.GetPolicy(ctx, req.GetProjectId())
	if err != nil { return nil, status.Error(codes.NotFound, err.Error()) }
	return &dbguardianv1.GetPolicyResponse{ProjectId: p.ProjectID, SpecYaml: p.SpecYAML, Version: int32(p.Version)}, nil
}

func (g *GuardianGRPCServer) UpdatePolicy(ctx context.Context, req *dbguardianv1.UpdatePolicyRequest) (*dbguardianv1.UpdatePolicyResponse, error) {
	if g.deps == nil || g.deps.PolicyMgr == nil { return nil, status.Error(codes.Unavailable, "policy manager unavailable") }
	p, err := g.deps.PolicyMgr.UpdatePolicy(ctx, req.GetProjectId(), req.GetSpecYaml())
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	return &dbguardianv1.UpdatePolicyResponse{ProjectId: p.ProjectID, SpecYaml: p.SpecYAML, Version: int32(p.Version)}, nil
}

func (g *GuardianGRPCServer) CheckDrift(ctx context.Context, req *dbguardianv1.CheckDriftRequest) (*dbguardianv1.CheckDriftResponse, error) {
	if g.deps == nil || g.deps.DriftCheck == nil { return nil, status.Error(codes.Unavailable, "drift check unavailable") }
	d, err := g.deps.DriftCheck.Check(ctx, req.GetProjectId())
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	return &dbguardianv1.CheckDriftResponse{MissingIndexes: d.MissingIndexes, ExtraIndexes: d.ExtraIndexes, RoleDrift: d.RoleDrift}, nil
}