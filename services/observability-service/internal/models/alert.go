package models

import (
	"fmt"
	"regexp"
	"time"
)

// AlertChannel represents notification channel types
type AlertChannel string

const (
	AlertChannelEmail     AlertChannel = "email"
	AlertChannelSlack     AlertChannel = "slack"
	AlertChannelPagerDuty AlertChannel = "pagerduty"
	AlertChannelWebhook   AlertChannel = "webhook"
)

// AlertRule represents an alert configuration linked to an SLO
type AlertRule struct {
	ID        string         `json:"id"`
	SLOID     string         `json:"slo_id"`
	Threshold float64        `json:"threshold"` // 0-100 percentage
	Channels  []AlertChannel `json:"channels"`
	Enabled   bool           `json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// UUID regex pattern
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Validate checks if alert rule has valid fields
func (a *AlertRule) Validate() error {
	if a.SLOID == "" {
		return fmt.Errorf("slo_id is required")
	}

	// Validate UUID format
	if !uuidRegex.MatchString(a.SLOID) {
		return fmt.Errorf("slo_id must be a valid UUID, got: %s", a.SLOID)
	}

	if a.Threshold < 0 || a.Threshold > 100 {
		return fmt.Errorf("threshold must be between 0 and 100, got: %f", a.Threshold)
	}

	if len(a.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}

	return nil
}

// AlertHistory represents an alert trigger record
type AlertHistory struct {
	ID          string    `json:"id"`
	AlertRuleID string    `json:"alert_rule_id"`
	SLOID       string    `json:"slo_id"`
	Timestamp   time.Time `json:"timestamp"`
	Compliance  float64   `json:"compliance"`
	Threshold   float64   `json:"threshold"`
	Notified    bool      `json:"notified"`
}

// ShouldNotify returns true if compliance below threshold and not yet notified
func (a *AlertHistory) ShouldNotify() bool {
	return a.Compliance < a.Threshold && !a.Notified
}
