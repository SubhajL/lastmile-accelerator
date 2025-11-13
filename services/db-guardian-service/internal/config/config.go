package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Environment  string
	ServiceName  string
	ServicePort  string
	LogLevel     string
	OTLPEndpoint string

	DatabaseURL string
	RedisAddr   string
	NATSUrl     string

	VaultAddr     string
	VaultRoleID   string
	VaultSecretID string

	// Analyzer defaults
	AnalysisMaxLockWarnTableSizeBytes int64
	AnalysisMinQueryExecutions        int
	AnalysisBenefitScoreThreshold     int

	// Auth (JWT/JWKS)
	AuthJWKSURL         string
	AuthIssuer          string
	AuthAudience        string
	AuthClockSkewSeconds int
}

func Load() (*Config, error) {
	cfg := &Config{
		Environment:   getEnv("ENV", "dev"),
		ServiceName:   os.Getenv("SERVICE_NAME"),
		ServicePort:   getEnv("SERVICE_PORT", "7105"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		OTLPEndpoint:  os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		NATSUrl:       os.Getenv("NATS_URL"),
		VaultAddr:     os.Getenv("VAULT_ADDR"),
		VaultRoleID:   os.Getenv("VAULT_ROLE_ID"),
		VaultSecretID: os.Getenv("VAULT_SECRET_ID"),
		// Analyzer defaults
		AnalysisMaxLockWarnTableSizeBytes: getEnvInt64("ANALYSIS_MAX_LOCK_WARN_TABLE_SIZE_BYTES", 50*1024*1024),
		AnalysisMinQueryExecutions:        getEnvInt("ANALYSIS_MIN_QUERY_EXECUTIONS", 100),
		AnalysisBenefitScoreThreshold:     getEnvInt("ANALYSIS_BENEFIT_SCORE_THRESHOLD", 10),
		// Auth
		AuthJWKSURL:          os.Getenv("AUTH_JWKS_URL"),
		AuthIssuer:           os.Getenv("AUTH_ISSUER"),
		AuthAudience:         os.Getenv("AUTH_AUDIENCE"),
		AuthClockSkewSeconds: getEnvInt("AUTH_CLOCK_SKEW_SECONDS", 60),
	}

	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("SERVICE_NAME is required")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	port, err := strconv.Atoi(c.ServicePort)
	if err != nil {
		return fmt.Errorf("invalid SERVICE_PORT '%s': must be numeric", c.ServicePort)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid SERVICE_PORT %d: must be between 1 and 65535", port)
	}

	if c.OTLPEndpoint != "" {
		if _, err := url.Parse(c.OTLPEndpoint); err != nil {
			return fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
		}
		// Additional check for valid scheme
		parsed, _ := url.Parse(c.OTLPEndpoint)
		if parsed.Scheme == "" {
			return fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: missing scheme")
		}
	}

	// Basic sanity checks for analyzer defaults
	if c.AnalysisMinQueryExecutions < 0 {
		return fmt.Errorf("ANALYSIS_MIN_QUERY_EXECUTIONS must be >= 0")
	}
	if c.AnalysisBenefitScoreThreshold < 0 {
		return fmt.Errorf("ANALYSIS_BENEFIT_SCORE_THRESHOLD must be >= 0")
	}
	if c.AnalysisMaxLockWarnTableSizeBytes < 0 {
		return fmt.Errorf("ANALYSIS_MAX_LOCK_WARN_TABLE_SIZE_BYTES must be >= 0")
	}

	// Auth config coherence: if any provided, require all and validate
	anyAuth := c.AuthJWKSURL != "" || c.AuthIssuer != "" || c.AuthAudience != ""
	if anyAuth {
		if c.AuthJWKSURL == "" || c.AuthIssuer == "" || c.AuthAudience == "" {
			return fmt.Errorf("auth config incomplete: require AUTH_JWKS_URL, AUTH_ISSUER, AUTH_AUDIENCE together")
		}
		u, err := url.Parse(c.AuthJWKSURL)
		if err != nil || u.Scheme == "" {
			return fmt.Errorf("invalid AUTH_JWKS_URL: %v", err)
		}
		if c.AuthClockSkewSeconds < 0 {
			return fmt.Errorf("AUTH_CLOCK_SKEW_SECONDS must be >= 0")
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return v
		}
	}
	return defaultValue
}
