package services

import (
	"context"
	"testing"

	"example.com/lma/observability-service/internal/models"
)

type fakeErrRepo struct{
	gid string
	groups []models.ErrorGroup
	got struct{ project, fp, title, msg, stack string }
}
func (f *fakeErrRepo) RecordEvent(ctx context.Context, projectID, fingerprint, title, message, stack string, metadata map[string]interface{}) (string, error){ f.gid = "g1"; f.got = struct{project,fp,title,msg,stack string}{projectID,fingerprint,title,message,stack}; return f.gid, nil }
func (f *fakeErrRepo) ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error){ return f.groups, nil }
func (f *fakeErrRepo) GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error){ if len(f.groups)>0 { return &f.groups[0], nil }; return nil, nil }
func (f *fakeErrRepo) ListEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error){ return nil, nil }
func (f *fakeErrRepo) ResolveGroup(ctx context.Context, groupID string) error{ return nil }

func TestErrorService_IngestAndList(t *testing.T){
	repo := &fakeErrRepo{}
	svc := NewErrorService(repo, nil)
	gid, err := svc.Ingest(context.Background(), "p1", "boom", "stack", "fp", "title", map[string]interface{}{"k":"v"})
	if err != nil || gid == "" { t.Fatalf("ingest failed: %v", err) }
	if repo.got.project != "p1" || repo.got.fp != "fp" { t.Fatalf("wrong record args") }
	_, _ = svc.ListGroups(context.Background(), "p1", models.ErrorGroupFilter{})
}
