package service

import (
	"context"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/repository"
)

// RecommendationsService combines stored index recommendations with ad-hoc role analysis.
type RecommendationsService struct {
	idxRepo     *repository.IndexRecommendationsRepository
	roleAnalyzer *analyzer.RoleAnalyzer
}

func NewRecommendationsService(idxRepo *repository.IndexRecommendationsRepository, roleAnalyzer *analyzer.RoleAnalyzer) *RecommendationsService {
	return &RecommendationsService{idxRepo: idxRepo, roleAnalyzer: roleAnalyzer}
}

func (s *RecommendationsService) Get(ctx context.Context, projectID string, onlyUnapplied bool) ([]analyzer.IndexRecommendation, *analyzer.RoleAnalysisResult, error) {
	// Stored index recommendations
	recs, err := s.idxRepo.ListByProject(ctx, projectID, onlyUnapplied)
	if err != nil { return nil, nil, err }
	var out []analyzer.IndexRecommendation
	for _, r := range recs {
		out = append(out, analyzer.IndexRecommendation{TableName: r.TableName, Columns: r.ColumnNames, Reason: r.Reason, BenefitScore: r.BenefitScore})
	}
	// Ad-hoc role analysis
	roleRes, err := s.roleAnalyzer.Analyze(ctx, analyzer.AnalyzeOptions{})
	if err != nil { return out, nil, err }
	return out, roleRes, nil
}

// PolicyService provides get/update around role policies
 type PolicyService struct { repo *repository.RolePoliciesRepository }

func NewPolicyService(repo *repository.RolePoliciesRepository) *PolicyService { return &PolicyService{repo: repo} }

func (s *PolicyService) GetPolicy(ctx context.Context, projectID string) (*models.RolePolicy, error) {
	p, err := s.repo.GetByProject(ctx, projectID)
	if err != nil { return nil, err }
	return p, nil
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, projectID, yaml string) (*models.RolePolicy, error) {
	m := &models.RolePolicy{ProjectID: projectID, SpecYAML: yaml}
	if _, err := s.repo.Upsert(ctx, m); err != nil { return nil, err }
	return m, nil
}
