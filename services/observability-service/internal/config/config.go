package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// Config holds all configuration for observability-service
type Config struct {
	Env             string
	ServiceName     string
	ServicePort     string
	GRPCPort        string
	DatabaseURL     string
	RedisURL        string
	NATSURL         string
	OTelEndpoint    string
	JWKSURL         string
	JWTIssuer       string
	JWTAudience     string // optional
	VaultAddr       string
	VaultRoleID     string
	VaultSecretID   string
	PrometheusURL   string
	TempoURL        string
	LokiURL         string
	SLOEvalInterval string // optional duration, default 1m
	SLOEvalWorkers  string // optional int, default 4
}

// Load reads configuration from environment variables and validates
func Load() (*Config, error) {
cfg := &Config{
		Env:             getEnv("ENV", "dev"),
		ServiceName:     getEnv("SERVICE_NAME", "observability-service"),
		ServicePort:     getEnv("SERVICE_PORT", ""),
		GRPCPort:        getEnv("GRPC_PORT", ""),
		DatabaseURL:     getEnv("DB_URL", ""),
		RedisURL:        getEnv("REDIS_URL", ""),
		NATSURL:         getEnv("NATS_URL", ""),
		OTelEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		JWKSURL:         getEnv("JWKS_URL", ""),
		JWTIssuer:       getEnv("JWT_ISSUER", ""),
		JWTAudience:     getEnv("JWT_AUDIENCE", ""),
		VaultAddr:       getEnv("VAULT_ADDR", ""),
		VaultRoleID:     getEnv("VAULT_ROLE_ID", ""),
		VaultSecretID:   getEnv("VAULT_SECRET_ID", ""),
		PrometheusURL:   getEnv("PROMETHEUS_URL", ""),
		TempoURL:        getEnv("TEMPO_URL", ""),
		LokiURL:         getEnv("LOKI_URL", ""),
		SLOEvalInterval: getEnv("SLO_EVAL_INTERVAL", ""),
		SLOEvalWorkers:  getEnv("SLO_EVAL_WORKERS", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks configuration constraints
func (c *Config) Validate() error {
	if c.ServicePort == "" {
		return fmt.Errorf("SERVICE_PORT is required")
	}
	if err := validatePort(c.ServicePort); err != nil {
		return fmt.Errorf("invalid SERVICE_PORT: %w", err)
	}

	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT is required")
	}
	if err := validatePort(c.GRPCPort); err != nil {
		return fmt.Errorf("invalid GRPC_PORT: %w", err)
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DB_URL is required")
	}

	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}

	if c.NATSURL == "" {
		return fmt.Errorf("NATS_URL is required")
	}

	if c.OTelEndpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required")
	}
	parsedURL, err := url.Parse(c.OTelEndpoint)
	if err != nil {
		return fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must have http or https scheme")
	}

if c.JWKSURL == "" {
		return fmt.Errorf("JWKS_URL is required")
	}
	if c.JWTIssuer == "" {
		return fmt.Errorf("JWT_ISSUER is required")
	}

	if c.PrometheusURL == "" {
		return fmt.Errorf("PROMETHEUS_URL is required")
	}
	if pu, err := url.Parse(c.PrometheusURL); err != nil || (pu.Scheme != "http" && pu.Scheme != "https") {
		return fmt.Errorf("invalid PROMETHEUS_URL")
	}

	if c.TempoURL == "" {
		return fmt.Errorf("TEMPO_URL is required")
	}
	if tu, err := url.Parse(c.TempoURL); err != nil || (tu.Scheme != "http" && tu.Scheme != "https") {
		return fmt.Errorf("invalid TEMPO_URL")
	}
	if c.LokiURL == "" {
		return fmt.Errorf("LOKI_URL is required")
	}
	if lu, err := url.Parse(c.LokiURL); err != nil || (lu.Scheme != "http" && lu.Scheme != "https") {
		return fmt.Errorf("invalid LOKI_URL")
	}

	if c.VaultAddr == "" {
		return fmt.Errorf("VAULT_ADDR is required")
	}

	if c.VaultRoleID == "" {
		return fmt.Errorf("VAULT_ROLE_ID is required")
	}

	if c.VaultSecretID == "" {
		return fmt.Errorf("VAULT_SECRET_ID is required")
	}

	return nil
}

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be a number: %w", err)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", p)
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
