package services

import (
	"context"
	"testing"

	"example.com/lma/observability-service/internal/models"
	"go.opentelemetry.io/otel/trace"
)

type fakeAlertRepo struct{ rules []*models.AlertRule; saved int }

func (f *fakeAlertRepo) Create(ctx context.Context, rule *models.AlertRule) error {
	f.rules = append(f.rules, rule)
	return nil
}
func (f *fakeAlertRepo) GetByID(ctx context.Context, id string) (*models.AlertRule, error) {
	for _, r := range f.rules {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}
func (f *fakeAlertRepo) GetBySLO(ctx context.Context, sloID string) ([]*models.AlertRule, error) {
	return f.rules, nil
}
func (f *fakeAlertRepo) Update(ctx context.Context, rule *models.AlertRule) error { return nil }
func (f *fakeAlertRepo) Delete(ctx context.Context, id string) error              { return nil }
func (f *fakeAlertRepo) SaveHistory(ctx context.Context, history *models.AlertHistory) error {
	f.saved++
	return nil
}
func (f *fakeAlertRepo) GetHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error) {
	return nil, nil
}

type fakePublisher struct{ called bool }

func (f *fakePublisher) PublishAlertTriggered(ctx context.Context, payload map[string]interface{}) error {
	f.called = true
	return nil
}

func TestAlertService_Evaluate_TriggersNotification(t *testing.T) {
	repo := &fakeAlertRepo{rules: []*models.AlertRule{{ID: "r1", SLOID: "s1", Threshold: 95.0, Channels: []models.AlertChannel{models.AlertChannelEmail}, Enabled: true}}}
	pub := &fakePublisher{}
	var noop trace.Tracer
	svc := NewAlertService(repo, noop, pub)

	status := &models.SLOStatus{SLOID: "s1", Compliance: 90.0}
	if err := svc.EvaluateAlerts(context.Background(), "s1", status); err != nil {
		t.Fatalf("EvaluateAlerts: %v", err)
	}
if !pub.called {
		t.Fatalf("expected publisher to be called")
}
if repo.saved == 0 {
				t.Fatalf("expected history to be saved")
}
}

func TestAlertService_Evaluate_NoTrigger_NoSaveHistory(t *testing.T) {
	repo := &fakeAlertRepo{rules: []*models.AlertRule{{ID: "r1", SLOID: "s1", Threshold: 80.0, Channels: []models.AlertChannel{models.AlertChannelEmail}, Enabled: true}}}
	pub := &fakePublisher{}
	var noop trace.Tracer
	svc := NewAlertService(repo, noop, pub)

	status := &models.SLOStatus{SLOID: "s1", Compliance: 95.0}
	if err := svc.EvaluateAlerts(context.Background(), "s1", status); err != nil {
		t.Fatalf("EvaluateAlerts: %v", err)
	}
	if repo.saved != 0 {
		t.Fatalf("did not expect history save")
	}
}
