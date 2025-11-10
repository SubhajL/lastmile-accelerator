package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/observability-service/internal/models"
)

type fakeErrSvc struct{
	gid string
	err error
}
func (f *fakeErrSvc) Ingest(ctx context.Context, projectID, message, stack, fingerprint, title string, metadata map[string]interface{}) (string, error){ return "g1", f.err }
func (f *fakeErrSvc) ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error){ return []models.ErrorGroup{{ID:"g1"}}, nil }
func (f *fakeErrSvc) GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error){ if groupID=="missing" { return nil, nil }; return &models.ErrorGroup{ID: groupID}, nil }
func (f *fakeErrSvc) ListGroupEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error){ return []models.ErrorEvent{{ID:"e1", GroupID: groupID}}, nil }
func (f *fakeErrSvc) ResolveGroup(ctx context.Context, groupID string) error { return nil }

func TestErrorHandler_Ingest_201(t *testing.T){
	svc := &fakeErrSvc{}
	h := NewErrorHandler(svc)
	payload := map[string]interface{}{"message":"boom","stack":"s","fingerprint":"fp","title":"t","metadata":map[string]interface{}{"k":"v"}}
	b,_ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/errors/ingest", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	h.Ingest(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("want 201 got %d", rec.Code) }
}

func TestErrorHandler_GetGroup_And_ListEvents_And_Resolve(t *testing.T){
	svc := &fakeErrSvc{}
	h := NewErrorHandler(svc)
	// GetGroup found
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/errors/groups/g1", nil)
	h.GetGroup(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec.Code) }
	// GetGroup not found
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/errors/groups/missing", nil)
	h.GetGroup(rec2, req2)
	if rec2.Code != http.StatusNotFound { t.Fatalf("want 404 got %d", rec2.Code) }
	// List events
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/v1/errors/groups/g1/events?limit=10&offset=0", nil)
	h.ListGroupEvents(rec3, req3)
	if rec3.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec3.Code) }
	// Resolve
	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/v1/errors/groups/g1/resolve", nil)
	h.ResolveGroup(rec4, req4)
	if rec4.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec4.Code) }
}

func TestErrorHandler_Ingest_TooLarge_413(t *testing.T){
	svc := &fakeErrSvc{}
	h := NewErrorHandler(svc)
big := bytes.Repeat([]byte("y"), 3<<20)
	body := []byte(`{"message":"`+string(big)+`"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/errors/ingest", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Ingest(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge { t.Fatalf("want 413 got %d", rec.Code) }
}
