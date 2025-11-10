package models

import (
	"testing"
	"time"
)

func TestSLO_Validate_Success(t *testing.T) {
	// Valid SLO passes all validations
	slo := &SLO{
		ProjectID:   "project-123",
		ServiceName: "api-service",
		Type:        SLOTypeAvailability,
		Target:      99.9,
		Window:      24 * time.Hour,
		Query:       "sum(rate(http_requests_total{status='200'}[5m]))",
	}

	if err := slo.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestSLO_Validate_InvalidTarget(t *testing.T) {
	// Target above 100 fails validation
	slo := &SLO{
		ProjectID:   "project-123",
		ServiceName: "api-service",
		Type:        SLOTypeAvailability,
		Target:      101.5,
		Window:      24 * time.Hour,
		Query:       "query",
	}

	err := slo.Validate()
	if err == nil {
		t.Fatal("expected validation error for target > 100")
	}
}

func TestSLO_Validate_NegativeWindow(t *testing.T) {
	// Negative window duration fails
	slo := &SLO{
		ProjectID:   "project-123",
		ServiceName: "api-service",
		Type:        SLOTypeLatency,
		Target:      95.0,
		Window:      -1 * time.Hour,
		Query:       "query",
	}

	err := slo.Validate()
	if err == nil {
		t.Fatal("expected validation error for negative window")
	}
}

func TestSLO_Validate_EmptyQuery(t *testing.T) {
	// Empty PromQL query fails
	slo := &SLO{
		ProjectID:   "project-123",
		ServiceName: "api-service",
		Type:        SLOTypeErrorRate,
		Target:      99.0,
		Window:      24 * time.Hour,
		Query:       "",
	}

	err := slo.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty query")
	}
}

func TestSLO_CalculateBudget_Availability(t *testing.T) {
	// 99.9% target returns 0.1% budget
	slo := &SLO{
		Type:   SLOTypeAvailability,
		Target: 99.9,
	}

	budget := slo.CalculateBudget()
	expected := 0.10000000000000142 // Floating point precision

	// Use delta comparison for floats
	delta := 0.0001
	if budget < expected-delta || budget > expected+delta {
		t.Errorf("expected budget ~%f, got %f", expected, budget)
	}
}

func TestSLOStatus_IsBreached_True(t *testing.T) {
	// Compliance below 100 returns true
	status := &SLOStatus{
		SLOID:      "slo-123",
		Compliance: 95.5,
	}

	if !status.IsBreached() {
		t.Error("expected status to be breached")
	}
}

func TestSLOStatus_IsBreached_False(t *testing.T) {
	// Compliance at 100 returns false
	status := &SLOStatus{
		SLOID:      "slo-123",
		Compliance: 100.0,
	}

	if status.IsBreached() {
		t.Error("expected status not to be breached")
	}
}

func TestSLOType_Valid(t *testing.T) {
	tests := []struct {
		sloType SLOType
		valid   bool
	}{
		{SLOTypeAvailability, true},
		{SLOTypeLatency, true},
		{SLOTypeErrorRate, true},
		{SLOTypeThroughput, true},
		{SLOType("invalid"), false},
	}

	for _, tt := range tests {
		result := IsValidSLOType(tt.sloType)
		if result != tt.valid {
			t.Errorf("IsValidSLOType(%v) = %v, want %v", tt.sloType, result, tt.valid)
		}
	}
}
