package service

import (
	"context"
	"sort"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/events"
	"example.com/lma/secrets-env-service/internal/repository"
)

type ParityService struct {
	parityRepo  *repository.ParityRepository
	secretsRepo repository.SecretsMetaRepo
	pub        events.Publisher
}

func NewParityService(parity *repository.ParityRepository, secrets repository.SecretsMetaRepo, pub events.Publisher) *ParityService {
	return &ParityService{parityRepo: parity, secretsRepo: secrets, pub: pub}
}

func (s *ParityService) CheckParity(ctx context.Context, projectID, baseEnv, compareEnv string) (*domain.EnvParityCheck, error) {
	// list base keys
	baseList, _, _ := s.secretsRepo.List(ctx, projectID, baseEnv, 1000, "")
	compareList, _, _ := s.secretsRepo.List(ctx, projectID, compareEnv, 1000, "")
	baseKeys := toKeys(baseList)
	compareKeys := toKeys(compareList)
	missing, extra := compareSecretKeys(baseKeys, compareKeys)
	check := &domain.EnvParityCheck{
		ProjectID: projectID,
		ScanTimestamp: time.Now(),
		MissingKeys: missing,
		ExtraKeys: extra,
	}
	_ = s.parityRepo.Create(ctx, check)
	if s.pub != nil {
		_ = s.pub.Publish(ctx, "parity.check.completed", map[string]any{"project_id": projectID, "base": baseEnv, "compare": compareEnv, "missing": len(missing), "extra": len(extra)})
		if check.HasDrift() {
			_ = s.pub.Publish(ctx, "env.parity.drift.detected", map[string]any{"project_id": projectID, "missing": missing, "extra": extra})
		}
	}
	return check, nil
}

func (s *ParityService) GetLatestCheck(ctx context.Context, projectID string) (*domain.EnvParityCheck, error) {
	return s.parityRepo.GetLatest(ctx, projectID)
}

func (s *ParityService) GetCheckHistory(ctx context.Context, projectID string, limit int) ([]*domain.EnvParityCheck, error) {
return s.parityRepo.GetHistory(ctx, projectID, limit)
}

func compareSecretKeys(base, compare []string) (missing, extra []string) {
	baseSet := make(map[string]struct{}, len(base))
	for _, k := range base { baseSet[k] = struct{}{} }
	cmpSet := make(map[string]struct{}, len(compare))
	for _, k := range compare { cmpSet[k] = struct{}{} }
	for k := range baseSet { if _, ok := cmpSet[k]; !ok { missing = append(missing, k) } }
	for k := range cmpSet { if _, ok := baseSet[k]; !ok { extra = append(extra, k) } }
	sort.Strings(missing); sort.Strings(extra)
	return
}

func toKeys(list []*domain.Secret) []string {
	out := make([]string, 0, len(list))
	for _, s := range list { out = append(out, s.Key) }
	return out
}
