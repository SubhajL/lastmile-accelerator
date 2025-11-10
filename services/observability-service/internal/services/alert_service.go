package services

import (
	"context"
	"fmt"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/repository"
	"go.opentelemetry.io/otel/trace"
)

type AlertPublisher interface {
	PublishAlertTriggered(ctx context.Context, payload map[string]interface{}) error
}

type AlertService interface {
	CreateAlert(ctx context.Context, rule *models.AlertRule) error
	GetAlert(ctx context.Context, id string) (*models.AlertRule, error)
	GetSLOAlerts(ctx context.Context, sloID string) ([]*models.AlertRule, error)
	UpdateAlert(ctx context.Context, rule *models.AlertRule) error
	DeleteAlert(ctx context.Context, id string) error
	EvaluateAlerts(ctx context.Context, sloID string, status *models.SLOStatus) error
	NotifyAlert(ctx context.Context, rule *models.AlertRule, status *models.SLOStatus) error
	GetAlertHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error)
}

type alertService struct {
	repo   repository.AlertRepository
	tracer trace.Tracer
	pub    AlertPublisher
}

func NewAlertService(repo repository.AlertRepository, tracer trace.Tracer, pub AlertPublisher) AlertService {
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("alert-service")
	}
	return &alertService{repo: repo, tracer: tracer, pub: pub}
}

func (s *alertService) CreateAlert(ctx context.Context, rule *models.AlertRule) error {
	return s.repo.Create(ctx, rule)
}
func (s *alertService) GetAlert(ctx context.Context, id string) (*models.AlertRule, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *alertService) GetSLOAlerts(ctx context.Context, sloID string) ([]*models.AlertRule, error) {
	return s.repo.GetBySLO(ctx, sloID)
}
func (s *alertService) UpdateAlert(ctx context.Context, rule *models.AlertRule) error {
	return s.repo.Update(ctx, rule)
}
func (s *alertService) DeleteAlert(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *alertService) EvaluateAlerts(ctx context.Context, sloID string, status *models.SLOStatus) error {
	rules, err := s.repo.GetBySLO(ctx, sloID)
	if err != nil {
		return err
	}
	for _, r := range rules {
		if s.shouldTrigger(r, status) {
			if err := s.NotifyAlert(ctx, r, status); err != nil {
				return err
			}
			// persist history when triggered
			h := &models.AlertHistory{ID: newUUID(), AlertRuleID: r.ID, SLOID: status.SLOID, Compliance: status.Compliance, Threshold: r.Threshold, Notified: true}
			if err := s.repo.SaveHistory(ctx, h); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *alertService) GetAlertHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error) {
	return s.repo.GetHistory(ctx, alertRuleID, limit)
}

func (s *alertService) NotifyAlert(ctx context.Context, rule *models.AlertRule, status *models.SLOStatus) error {
	payload := map[string]interface{}{
		"alert_id":   rule.ID,
		"slo_id":     status.SLOID,
		"compliance": status.Compliance,
		"threshold":  rule.Threshold,
		"channels":   rule.Channels,
	}
	if s.pub != nil {
		if err := s.pub.PublishAlertTriggered(ctx, payload); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}
	return nil
}

func (s *alertService) shouldTrigger(rule *models.AlertRule, status *models.SLOStatus) bool {
	if !rule.Enabled {
		return false
	}
	return status.Compliance < rule.Threshold
}
