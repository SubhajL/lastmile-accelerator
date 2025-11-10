package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

type AlertRepository interface {
	Create(ctx context.Context, rule *models.AlertRule) error
	GetByID(ctx context.Context, id string) (*models.AlertRule, error)
	GetBySLO(ctx context.Context, sloID string) ([]*models.AlertRule, error)
	Update(ctx context.Context, rule *models.AlertRule) error
	Delete(ctx context.Context, id string) error
	SaveHistory(ctx context.Context, history *models.AlertHistory) error
	GetHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error)
}

type postgresAlertRepo struct {
	db *storage.PostgresDB
}

func NewAlertRepository(db *storage.PostgresDB) AlertRepository {
	return &postgresAlertRepo{db: db}
}

func (r *postgresAlertRepo) Create(ctx context.Context, rule *models.AlertRule) error {
	channels := map[string][]string{"channels": {}}
	for _, ch := range rule.Channels {
		channels["channels"] = append(channels["channels"], string(ch))
	}
	b, _ := json.Marshal(channels)
	const q = `INSERT INTO alert_rules (id, slo_id, threshold, channels, enabled, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW())`
	_, err := r.db.DB().ExecContext(ctx, q, rule.ID, rule.SLOID, rule.Threshold, string(b), rule.Enabled)
	if err != nil {
		return fmt.Errorf("insert alert_rule: %w", err)
	}
	return nil
}

func (r *postgresAlertRepo) GetByID(ctx context.Context, id string) (*models.AlertRule, error) {
	const q = `SELECT id, slo_id, threshold, channels, enabled, created_at, updated_at FROM alert_rules WHERE id=$1`
	row := r.db.DB().QueryRowContext(ctx, q, id)
	var (
		rule models.AlertRule
		ch   JSONMap
	)
	if err := row.Scan(&rule.ID, &rule.SLOID, &rule.Threshold, &ch.Raw, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get alert_rule: %w", err)
	}
	rule.Channels = ch.ToChannels()
	return &rule, nil
}

func (r *postgresAlertRepo) GetBySLO(ctx context.Context, sloID string) ([]*models.AlertRule, error) {
	const q = `SELECT id, slo_id, threshold, channels, enabled, created_at, updated_at FROM alert_rules WHERE slo_id=$1`
	rows, err := r.db.DB().QueryContext(ctx, q, sloID)
	if err != nil {
		return nil, fmt.Errorf("get by slo: %w", err)
	}
	defer rows.Close()
	var out []*models.AlertRule
	for rows.Next() {
		var rule models.AlertRule
		var ch JSONMap
		if err := rows.Scan(&rule.ID, &rule.SLOID, &rule.Threshold, &ch.Raw, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rule.Channels = ch.ToChannels()
		out = append(out, &rule)
	}
	return out, nil
}

func (r *postgresAlertRepo) Update(ctx context.Context, rule *models.AlertRule) error {
	channels := map[string][]string{"channels": {}}
	for _, ch := range rule.Channels {
		channels["channels"] = append(channels["channels"], string(ch))
	}
	b, _ := json.Marshal(channels)
	const q = `UPDATE alert_rules SET slo_id=$2, threshold=$3, channels=$4, enabled=$5, updated_at=NOW() WHERE id=$1`
	_, err := r.db.DB().ExecContext(ctx, q, rule.ID, rule.SLOID, rule.Threshold, string(b), rule.Enabled)
	if err != nil {
		return fmt.Errorf("update alert_rule: %w", err)
	}
	return nil
}

func (r *postgresAlertRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM alert_rules WHERE id=$1`
	_, err := r.db.DB().ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete alert_rule: %w", err)
	}
	return nil
}

func (r *postgresAlertRepo) SaveHistory(ctx context.Context, history *models.AlertHistory) error {
	const q = `INSERT INTO alert_history (id, alert_rule_id, slo_id, timestamp, compliance, threshold, notified) VALUES ($1,$2,$3,NOW(),$4,$5,$6)`
	_, err := r.db.DB().ExecContext(ctx, q, history.ID, history.AlertRuleID, history.SLOID, history.Compliance, history.Threshold, history.Notified)
	if err != nil {
		return fmt.Errorf("insert alert_history: %w", err)
	}
	return nil
}

func (r *postgresAlertRepo) GetHistory(ctx context.Context, alertRuleID string, limit int) ([]*models.AlertHistory, error) {
	const q = `SELECT id, alert_rule_id, slo_id, timestamp, compliance, threshold, notified FROM alert_history WHERE alert_rule_id=$1 ORDER BY timestamp DESC LIMIT 100`
	rows, err := r.db.DB().QueryContext(ctx, q, alertRuleID)
	if err != nil {
		return nil, fmt.Errorf("get alert_history: %w", err)
	}
	defer rows.Close()
	var out []*models.AlertHistory
	for rows.Next() {
		var h models.AlertHistory
		if err := rows.Scan(&h.ID, &h.AlertRuleID, &h.SLOID, &h.Timestamp, &h.Compliance, &h.Threshold, &h.Notified); err != nil {
			return nil, err
		}
		out = append(out, &h)
	}
	return out, nil
}

// JSONMap helps scan JSON string and convert to channels
type JSONMap struct{ Raw string }

func (j JSONMap) ToChannels() []models.AlertChannel {
	m := map[string][]string{}
	_ = json.Unmarshal([]byte(j.Raw), &m)
	var out []models.AlertChannel
	for _, s := range m["channels"] {
		out = append(out, models.AlertChannel(s))
	}
	return out
}
