package models

import (
	"fmt"
	"time"
)

// SLOType represents the type of SLO
type SLOType string

const (
	SLOTypeAvailability SLOType = "availability"
	SLOTypeLatency      SLOType = "latency"
	SLOTypeErrorRate    SLOType = "error_rate"
	SLOTypeThroughput   SLOType = "throughput"
)

// IsValidSLOType checks if SLO type is supported
func IsValidSLOType(t SLOType) bool {
	switch t {
	case SLOTypeAvailability, SLOTypeLatency, SLOTypeErrorRate, SLOTypeThroughput:
		return true
	}
	return false
}

// SLO represents a Service Level Objective definition
type SLO struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	ServiceName string        `json:"service_name"`
	Type        SLOType       `json:"type"`
	Target      float64       `json:"target"` // 0-100 percentage
	Window      time.Duration `json:"window"` // Time window for SLO
	Query       string        `json:"query"`  // PromQL query
	Status      string        `json:"status"` // active, paused, etc.
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Validate checks if SLO has valid fields
func (s *SLO) Validate() error {
	if s.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	if s.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if !IsValidSLOType(s.Type) {
		return fmt.Errorf("invalid SLO type: %s", s.Type)
	}

	if s.Target < 0 || s.Target > 100 {
		return fmt.Errorf("target must be between 0 and 100, got: %f", s.Target)
	}

	if s.Window <= 0 {
		return fmt.Errorf("window must be positive, got: %s", s.Window)
	}

	if s.Query == "" {
		return fmt.Errorf("query is required")
	}

	return nil
}

// CalculateBudget computes error budget as (100 - Target)
func (s *SLO) CalculateBudget() float64 {
	return 100 - s.Target
}

// SLOStatus represents current SLO compliance status
type SLOStatus struct {
	SLOID           string    `json:"slo_id"`
	Compliance      float64   `json:"compliance"`       // 0-100 percentage
	BurnRate        float64   `json:"burn_rate"`        // Rate of budget consumption
	RemainingBudget float64   `json:"remaining_budget"` // Remaining error budget
	LastCalculated  time.Time `json:"last_calculated"`
}

// IsBreached returns true if compliance is below target (100%)
func (s *SLOStatus) IsBreached() bool {
	return s.Compliance < 100
}

// SLOHistory represents historical compliance record
type SLOHistory struct {
	SLOID      string    `json:"slo_id"`
	Timestamp  time.Time `json:"timestamp"`
	Compliance float64   `json:"compliance"`
	BurnRate   float64   `json:"burn_rate"`
}
