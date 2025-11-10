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

type fakeAlertService struct{
	created *models.AlertRule
	rule *models.AlertRule
	rules []*models.AlertRule
	hist []*models.AlertHistory
	err error
}

func (f *fakeAlertService) CreateAlert(ctx context.Context, rule *models.AlertRule) error { f.created = rule; return f.err }
func (f *fakeAlertService) GetAlert(ctx context.Context, id string) (*models.AlertRule, error) { if f.rule==nil { return nil, f.err }; return f.rule, nil }
func (f *fakeAlertService) GetSLOAlerts(ctx context.Context, sloID string) ([]*models.AlertRule, error) { return f.rules, nil }
func (f *fakeAlertService) UpdateAlert(ctx context.Context, rule *models.AlertRule) error { return f.err }
func (f *fakeAlertService) DeleteAlert(ctx context.Context, id string) error { return f.err }
func (f *fakeAlertService) EvaluateAlerts(ctx context.Context, sloID string, status *models.SLOStatus) error { return nil }
func (f *fakeAlertService) NotifyAlert(ctx context.Context, rule *models.AlertRule, status *models.SLOStatus) error { return nil }
func (f *fakeAlertService) GetAlertHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error) { return f.hist, nil }

func TestAlertHandler_CreateAndList(t *testing.T){
	svc := &fakeAlertService{}
	h := NewAlertHandler(svc)
var reqPayload = struct{
		Threshold float64 `json:"threshold"`
		Channels []string `json:"channels"`
		Enabled bool `json:"enabled"`
	}{Threshold:95.0, Channels:[]string{"email"}, Enabled:true}
	body,_ := json.Marshal(reqPayload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/slos/s1/alerts", bytes.NewReader(body))
	h.CreateAlert(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("want 201 got %d", rec.Code) }

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/slos/s1/alerts", nil)
	h.GetSLOAlerts(rec2, req2)
	if rec2.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec2.Code) }
}

func TestAlertHandler_Create_UnknownField_400(t *testing.T){
	svc := &fakeAlertService{}
	h := NewAlertHandler(svc)
	rec := httptest.NewRecorder()
	body := []byte(`{"threshold":1.0,"channels":["email"],"enabled":true,"extra":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/slos/s1/alerts", bytes.NewReader(body))
	h.CreateAlert(rec, req)
	if rec.Code != http.StatusBadRequest { t.Fatalf("want 400 got %d", rec.Code) }
}

func TestAlertHandler_Create_TooLarge_413(t *testing.T){
	svc := &fakeAlertService{}
	h := NewAlertHandler(svc)
	rec := httptest.NewRecorder()
big := bytes.Repeat([]byte("x"), 2<<20)
	body := []byte(`{"threshold":1.0,"channels":["`+string(big)+`"],"enabled":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/slos/s1/alerts", bytes.NewReader(body))
	h.CreateAlert(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge { t.Fatalf("want 413 got %d", rec.Code) }
}

func TestAlertHandler_GetPutDelete_History(t *testing.T){
	svc := &fakeAlertService{rule:&models.AlertRule{ID:"r1", SLOID:"s1"}, hist: []*models.AlertHistory{{ID:"h1", AlertRuleID:"r1", SLOID:"s1", Notified:true}}}
	h := NewAlertHandler(svc)
	// Get
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/alerts/r1", nil)
	h.GetAlert(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec.Code) }
// Update
var updPayload = struct{
	Threshold float64 `json:"threshold"`
	Channels []string `json:"channels"`
	Enabled bool `json:"enabled"`
}{Threshold:90.0, Channels:[]string{"slack"}, Enabled:false}
body,_ := json.Marshal(updPayload)
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPut, "/v1/alerts/r1", bytes.NewReader(body))
	h.UpdateAlert(rec2, req2)
	if rec2.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec2.Code) }
	// Delete
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, "/v1/alerts/r1", nil)
	h.DeleteAlert(rec3, req3)
	if rec3.Code != http.StatusNoContent { t.Fatalf("want 204 got %d", rec3.Code) }
	// History
	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/v1/alerts/r1/history?limit=10", nil)
	h.GetAlertHistory(rec4, req4)
	if rec4.Code != http.StatusOK { t.Fatalf("want 200 got %d", rec4.Code) }
}
