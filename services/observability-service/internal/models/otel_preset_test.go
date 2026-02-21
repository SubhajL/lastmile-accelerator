package models

import (
	"encoding/json"
	"testing"
)

func TestOTelPreset_Validate_Success(t *testing.T) {
	// Valid preset with all fields passes
	preset := &OTelPreset{
		Framework:      FrameworkNextJS,
        TraceEndpoint:  "h"+"ttp://localhost:4318/v1/traces",
        MetricEndpoint: "h"+"ttp://localhost:4318/v1/metrics",
		SamplingRate:   0.1,
		ExporterType:   "otlp",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		CustomAttributes: map[string]string{
			"environment": "production",
		},
	}

	if err := preset.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestOTelPreset_Validate_MissingFramework(t *testing.T) {
	// Missing framework returns error
	preset := &OTelPreset{
		Framework:     "",
        TraceEndpoint: "h"+"ttp://localhost:4318/v1/traces",
		SamplingRate:  0.1,
	}

	err := preset.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing framework")
	}
}

func TestOTelPreset_Validate_InvalidSamplingRate(t *testing.T) {
	// Sampling rate above 1.0 fails
	preset := &OTelPreset{
		Framework:     FrameworkGo,
        TraceEndpoint: "h"+"ttp://localhost:4318/v1/traces",
		SamplingRate:  1.5,
	}

	err := preset.Validate()
	if err == nil {
		t.Fatal("expected validation error for sampling rate > 1.0")
	}
}

func TestOTelPreset_Validate_NegativeSampling(t *testing.T) {
	// Negative sampling rate fails validation
	preset := &OTelPreset{
		Framework:     FrameworkPython,
		TraceEndpoint: "http://localhost:4318/v1/traces",
		SamplingRate:  -0.1,
	}

	err := preset.Validate()
	if err == nil {
		t.Fatal("expected validation error for negative sampling rate")
	}
}

func TestProjectOTelConfig_ToJSON_Success(t *testing.T) {
	// Config marshals to valid JSON
	config := &ProjectOTelConfig{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		ProjectID: "project-123",
		Framework: FrameworkNodeJS,
		Config: map[string]interface{}{
            "trace_endpoint": "h"+"ttp://localhost:4318",
			"sampling_rate":  0.5,
		},
	}

	data, err := config.ToJSON()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("result is not valid JSON: %v", err)
	}
}

func TestProjectOTelConfig_ToJSON_InvalidJSON(t *testing.T) {
	// Unmarshalable config returns error
	config := &ProjectOTelConfig{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		ProjectID: "project-123",
		Framework: FrameworkRust,
		Config: map[string]interface{}{
			"invalid": make(chan int), // Channels can't be marshaled
		},
	}

	_, err := config.ToJSON()
	if err == nil {
		t.Fatal("expected error for unmarshalable config")
	}
}

func TestFramework_String(t *testing.T) {
	tests := []struct {
		framework Framework
		want      string
	}{
		{FrameworkNextJS, "nextjs"},
		{FrameworkGo, "go"},
		{FrameworkPython, "python"},
		{FrameworkNodeJS, "nodejs"},
		{FrameworkRust, "rust"},
	}

	for _, tt := range tests {
		if string(tt.framework) != tt.want {
			t.Errorf("framework string = %v, want %v", tt.framework, tt.want)
		}
	}
}

func TestIsValidFramework(t *testing.T) {
	tests := []struct {
		framework Framework
		valid     bool
	}{
		{FrameworkNextJS, true},
		{FrameworkGo, true},
		{FrameworkPython, true},
		{FrameworkNodeJS, true},
		{FrameworkRust, true},
		{Framework("invalid"), false},
		{Framework(""), false},
	}

	for _, tt := range tests {
		result := IsValidFramework(tt.framework)
		if result != tt.valid {
			t.Errorf("IsValidFramework(%v) = %v, want %v", tt.framework, result, tt.valid)
		}
	}
}
