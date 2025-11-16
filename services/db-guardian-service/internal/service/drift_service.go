package service

import (
	"context"
	"strings"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/dto"
	"example.com/lma/db-guardian-service/internal/repository"
)

// DriftService compares live DB with stored artifacts.
type DriftService struct {
	inspector analyzer.DBInspector
	idxRepo   *repository.IndexRecommendationsRepository
}

func NewDriftService(inspector analyzer.DBInspector, idxRepo *repository.IndexRecommendationsRepository) *DriftService {
	return &DriftService{inspector: inspector, idxRepo: idxRepo}
}

func (s *DriftService) Check(ctx context.Context, projectID string) (*dto.DriftResponse, error) {
	// Very simple index drift: check presence of recommended indexes by name pattern
	recs, err := s.idxRepo.ListByProject(ctx, projectID, true)
	if err != nil { return nil, err }

	missing, extra := []string{}, []string{}
	// Build set of recommended index names for quick matching
	recNames := map[string]bool{}
	for _, r := range recs {
		name := "idx_" + strings.Join(append([]string{r.TableName}, r.ColumnNames...), "_")
		recNames[name] = true
	}
	// Fetch live indexes per table in recs
	tables := map[string]bool{}
	for _, r := range recs { tables[r.TableName] = true }
	for tbl := range tables {
		idxs, err := s.inspector.GetIndexes(ctx, tbl)
		if err != nil { continue }
		live := map[string]bool{}
		for _, i := range idxs { live[i.Name] = true }
        for rn := range recNames { if strings.HasPrefix(rn, "idx_"+tbl) && !live[rn] { missing = append(missing, rn) } }
        for ln := range live { if !recNames[ln] { extra = append(extra, ln) } }
	}
	return &dto.DriftResponse{MissingIndexes: missing, ExtraIndexes: extra, RoleDrift: []string{}}, nil
}
