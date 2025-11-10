package models

import (
	"testing"
)

func TestAlertRule_Validate_Success(t *testing.T) {
	// Valid alert passes validation
	alert := &AlertRule{
		SLOID:     "550e8400-e29b-41d4-a716-446655440000",
		Threshold: 95.0,
		Channels:  []AlertChannel{AlertChannelEmail, AlertChannelSlack},
		Enabled:   true,
	}

	if err := alert.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestAlertRule_Validate_InvalidThreshold(t *testing.T) {
	// Threshold above 100 fails
	alert := &AlertRule{
		SLOID:     "550e8400-e29b-41d4-a716-446655440000",
		Threshold: 105.0,
		Channels:  []AlertChannel{AlertChannelEmail},
	}

	err := alert.Validate()
	if err == nil {
		t.Fatal("expected validation error for threshold > 100")
	}
}

func TestAlertRule_Validate_NoChannels(t *testing.T) {
	// Empty channels array fails
	alert := &AlertRule{
		SLOID:     "550e8400-e29b-41d4-a716-446655440000",
		Threshold: 95.0,
		Channels:  []AlertChannel{},
	}

	err := alert.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty channels")
	}
}

func TestAlertRule_Validate_InvalidSLOID(t *testing.T) {
	// Malformed UUID fails validation
	alert := &AlertRule{
		SLOID:     "not-a-uuid",
		Threshold: 95.0,
		Channels:  []AlertChannel{AlertChannelPagerDuty},
	}

	err := alert.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid UUID")
	}
}

func TestAlertHistory_ShouldNotify_True(t *testing.T) {
	// Breach with notified false returns true
	history := &AlertHistory{
		AlertRuleID: "rule-123",
		Compliance:  85.0,
		Threshold:   95.0,
		Notified:    false,
	}

	if !history.ShouldNotify() {
		t.Error("expected ShouldNotify to return true")
	}
}

func TestAlertHistory_ShouldNotify_False(t *testing.T) {
	// Already notified returns false
	history := &AlertHistory{
		AlertRuleID: "rule-123",
		Compliance:  85.0,
		Threshold:   95.0,
		Notified:    true,
	}

	if history.ShouldNotify() {
		t.Error("expected ShouldNotify to return false when already notified")
	}
}

func TestAlertHistory_ShouldNotify_AboveThreshold(t *testing.T) {
	// Compliance above threshold returns false
	history := &AlertHistory{
		AlertRuleID: "rule-123",
		Compliance:  98.0,
		Threshold:   95.0,
		Notified:    false,
	}

	if history.ShouldNotify() {
		t.Error("expected ShouldNotify to return false when above threshold")
	}
}

func TestAlertChannel_Valid(t *testing.T) {
	channels := []AlertChannel{
		AlertChannelEmail,
		AlertChannelSlack,
		AlertChannelPagerDuty,
		AlertChannelWebhook,
	}

	for _, ch := range channels {
		if ch == "" {
			t.Errorf("channel should not be empty: %v", ch)
		}
	}
}
