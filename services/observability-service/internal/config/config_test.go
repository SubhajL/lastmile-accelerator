package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Valid env vars produce complete config
	setEnv := func(key, value string) {
		t.Setenv(key, value)
	}

	setEnv("ENV", "dev")
	setEnv("SERVICE_NAME", "observability-service")
	setEnv("SERVICE_PORT", "7301")
	setEnv("GRPC_PORT", "50081")
	setEnv("DB_URL", "postgres://user:pass@localhost:5432/obs")
	setEnv("REDIS_URL", "redis://localhost:6379")
	setEnv("NATS_URL", "nats://localhost:4222")
	setEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
setEnv("JWKS_URL", "https://auth.example.com/.well-known/jwks.json")
setEnv("JWT_ISSUER", "https://auth.example.com/")
setEnv("TEMPO_URL", "http://localhost:3200")
setEnv("LOKI_URL", "http://localhost:3100")
	setEnv("VAULT_ADDR", "http://localhost:8200")
	setEnv("VAULT_ROLE_ID", "test-role")
	setEnv("VAULT_SECRET_ID", "test-secret")
	setEnv("PROMETHEUS_URL", "http://localhost:9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.ServicePort != "7301" {
		t.Errorf("expected ServicePort 7301, got %s", cfg.ServicePort)
	}
	if cfg.GRPCPort != "50081" {
		t.Errorf("expected GRPCPort 50081, got %s", cfg.GRPCPort)
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected DatabaseURL to be set")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	// Missing SERVICE_PORT returns validation error
	os.Clearenv()
	t.Setenv("ENV", "dev")
	t.Setenv("SERVICE_NAME", "observability-service")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for missing SERVICE_PORT")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	// Port outside 1-65535 fails validation
	t.Setenv("ENV", "dev")
	t.Setenv("SERVICE_NAME", "observability-service")
	t.Setenv("SERVICE_PORT", "99999")
	t.Setenv("GRPC_PORT", "50081")
	t.Setenv("DB_URL", "postgres://localhost/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("NATS_URL", "nats://localhost:4222")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
t.Setenv("JWKS_URL", "https://auth.example.com/jwks.json")
t.Setenv("JWT_ISSUER", "https://auth.example.com/")
t.Setenv("TEMPO_URL", "http://localhost:3200")
t.Setenv("LOKI_URL", "http://localhost:3100")
	t.Setenv("VAULT_ADDR", "http://localhost:8200")
	t.Setenv("VAULT_ROLE_ID", "test-role")
	t.Setenv("VAULT_SECRET_ID", "test-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid port")
	}
}

func TestValidate_Success(t *testing.T) {
	// All fields valid passes validation
	cfg := &Config{
		Env:             "dev",
		ServiceName:     "observability-service",
		ServicePort:     "7301",
		GRPCPort:        "50081",
		DatabaseURL:     "postgres://localhost/db",
		RedisURL:        "redis://localhost:6379",
		NATSURL:         "nats://localhost:4222",
		OTelEndpoint:    "http://localhost:4317",
JWKSURL: "https://auth.example.com/jwks.json",
JWTIssuer: "https://auth.example.com/",
TempoURL: "http://localhost:3200",
LokiURL: "http://localhost:3100",
		VaultAddr:       "http://localhost:8200",
		VaultRoleID:     "role",
		VaultSecretID: "secret",
		PrometheusURL: "http://localhost:9090",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestValidate_InvalidOTelEndpoint(t *testing.T) {
	// Also validate invalid Prometheus URL
	// Malformed OTEL URL returns specific error
	cfg := &Config{
		Env:             "dev",
		ServiceName:     "observability-service",
		ServicePort:     "7301",
		GRPCPort:        "50081",
		DatabaseURL:     "postgres://localhost/db",
		RedisURL:        "redis://localhost:6379",
		NATSURL:         "nats://localhost:4222",
		OTelEndpoint:    "not-a-url",
JWKSURL: "https://auth.example.com/jwks.json",
JWTIssuer: "https://auth.example.com/",
TempoURL: "http://localhost:3200",
LokiURL: "http://localhost:3100",
		VaultAddr:       "http://localhost:8200",
		VaultRoleID:     "role",
		VaultSecretID: "secret",
		PrometheusURL: "not-a-url",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid OTel endpoint")
	}
}
