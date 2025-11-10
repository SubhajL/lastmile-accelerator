package config

import (
	"os"
	"testing"
)

func TestLoad_WithAllEnvVars_ReturnsValidConfig(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"ENV":                          "staging",
		"SERVICE_NAME":                 "db-guardian-service",
		"SERVICE_PORT":                 "7105",
		"OTEL_EXPORTER_OTLP_ENDPOINT":  "http://otel-collector:4317",
		"LOG_LEVEL":                    "info",
		"DATABASE_URL":                 "postgres://user:pass@localhost:5432/dbname",
		"REDIS_ADDR":                   "localhost:6379",
		"NATS_URL":                     "nats://localhost:4222",
		"VAULT_ADDR":                   "http://vault:8200",
		"VAULT_ROLE_ID":                "test-role",
		"VAULT_SECRET_ID":              "test-secret",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ServiceName != "db-guardian-service" {
		t.Errorf("expected service name 'db-guardian-service', got '%s'", cfg.ServiceName)
	}
	if cfg.ServicePort != "7105" {
		t.Errorf("expected port '7105', got '%s'", cfg.ServicePort)
	}
	if cfg.Environment != "staging" {
		t.Errorf("expected environment 'staging', got '%s'", cfg.Environment)
	}
}

func TestLoad_WithDefaults_UsesDefaultValues(t *testing.T) {
	// Arrange
	os.Setenv("SERVICE_NAME", "db-guardian-service")
	defer os.Unsetenv("SERVICE_NAME")

	// Act
	cfg, err := Load()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ServicePort != "7105" {
		t.Errorf("expected default port '7105', got '%s'", cfg.ServicePort)
	}
	if cfg.Environment != "dev" {
		t.Errorf("expected default environment 'dev', got '%s'", cfg.Environment)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.LogLevel)
	}
}

func TestLoad_MissingRequired_ReturnsError(t *testing.T) {
	// Arrange - no SERVICE_NAME set
	os.Unsetenv("SERVICE_NAME")

	// Act
	_, err := Load()

	// Assert
	if err == nil {
		t.Fatal("expected error for missing SERVICE_NAME, got nil")
	}
}

func TestValidate_InvalidPort_ReturnsError(t *testing.T) {
	// Arrange
	cfg := &Config{
		ServiceName: "test",
		ServicePort: "99999", // Invalid port
		Environment: "dev",
	}

	// Act
	err := cfg.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestValidate_InvalidOTLPEndpoint_ReturnsError(t *testing.T) {
	// Arrange
	cfg := &Config{
		ServiceName:    "test",
		ServicePort:    "7105",
		Environment:    "dev",
		OTLPEndpoint:   "not-a-url",
	}

	// Act
	err := cfg.Validate()

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid OTLP endpoint, got nil")
	}
}
