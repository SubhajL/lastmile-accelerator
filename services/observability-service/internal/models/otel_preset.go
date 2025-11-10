package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Framework represents supported framework types
type Framework string

const (
	FrameworkNextJS Framework = "nextjs"
	FrameworkGo     Framework = "go"
	FrameworkPython Framework = "python"
	FrameworkNodeJS Framework = "nodejs"
	FrameworkRust   Framework = "rust"
)

// IsValidFramework checks if framework is supported
func IsValidFramework(f Framework) bool {
	switch f {
	case FrameworkNextJS, FrameworkGo, FrameworkPython, FrameworkNodeJS, FrameworkRust:
		return true
	}
	return false
}

// OTelPreset represents a framework-specific OTel configuration
type OTelPreset struct {
	Framework        Framework         `json:"framework"`
	TraceEndpoint    string            `json:"trace_endpoint"`
	MetricEndpoint   string            `json:"metric_endpoint"`
	SamplingRate     float64           `json:"sampling_rate"`
	ExporterType     string            `json:"exporter_type"`
	Headers          map[string]string `json:"headers,omitempty"`
	CustomAttributes map[string]string `json:"custom_attributes,omitempty"`
}

// Validate checks if preset has required fields and valid values
func (o *OTelPreset) Validate() error {
	if o.Framework == "" {
		return fmt.Errorf("framework is required")
	}

	if !IsValidFramework(o.Framework) {
		return fmt.Errorf("unsupported framework: %s", o.Framework)
	}

	if o.TraceEndpoint == "" {
		return fmt.Errorf("trace_endpoint is required")
	}

	if o.SamplingRate < 0 || o.SamplingRate > 1 {
		return fmt.Errorf("sampling_rate must be between 0 and 1, got: %f", o.SamplingRate)
	}

	return nil
}

// ProjectOTelConfig stores project-applied OTel configuration
type ProjectOTelConfig struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Framework Framework              `json:"framework"`
	Config    map[string]interface{} `json:"config"`
	AppliedAt time.Time              `json:"applied_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ToJSON marshals Config field to JSON
func (p *ProjectOTelConfig) ToJSON() ([]byte, error) {
	data, err := json.Marshal(p.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	return data, nil
}
