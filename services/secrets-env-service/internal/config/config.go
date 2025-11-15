package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the service
type Config struct {
	Env           string
	ServiceName   string
	ServicePort   string
	LogLevel      string
	Vault         VaultConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	NATS          NATSConfig
	Observability ObservabilityConfig
	Auth          AuthConfig
	Storage       StorageConfig
	TLS           TLSConfig
	RateLimit     RateLimitConfig
    AllowedEnvironments []string
}

// VaultConfig holds Vault-specific configuration
type VaultConfig struct {
	Address   string
	RoleID    string
	SecretID  string
	Namespace string
	MountPath string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	URL              string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetimeS int
	UsePGRepos       bool
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL string
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string
}

// ObservabilityConfig holds observability settings
type ObservabilityConfig struct {
    OTELEndpoint   string
    OTELInsecure   bool
    OTELHeaders    map[string]string
    OTELServiceName string
    MetricsPort    string
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTPublicKey string
	JWKSURL      string
	Audience     string
	Issuer       string
	JWKSCacheTTL int // seconds
}

// StorageConfig holds storage-related settings
type StorageConfig struct {
	S3 S3Config
}

// TLSConfig holds optional server-side TLS settings (mTLS).
type TLSConfig struct {
	EnableMTLS   bool
	CertFile     string
	KeyFile      string
	ClientCAFile string
}

// RateLimitConfig configures global rate-limiting defaults.
type RateLimitConfig struct {
	RequestsPerSec int
	Burst          int
}

// S3Config holds S3 storage settings
type S3Config struct {
	Endpoint       string
	Bucket         string
	Prefix         string
	AccessKey      string
	SecretKey      string
	UseTLS         bool
	IgnoreGlobs    []string
	SizeLimitBytes int64
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Env:         getEnvOrDefault("ENV", "dev"),
		ServiceName: getEnvOrDefault("SERVICE_NAME", "secrets-env-service"),
		ServicePort: getEnvOrDefault("SERVICE_PORT", "7104"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
		Vault: VaultConfig{
			Address:   os.Getenv("VAULT_ADDR"),
			RoleID:    os.Getenv("VAULT_ROLE_ID"),
			SecretID:  os.Getenv("VAULT_SECRET_ID"),
			Namespace: getEnvOrDefault("VAULT_NAMESPACE", ""),
			MountPath: getEnvOrDefault("VAULT_MOUNT_PATH", "secret"),
		},
		Database: DatabaseConfig{
			URL:              os.Getenv("DATABASE_URL"),
			MaxOpenConns:     getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:     getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeS: getEnvAsInt("DB_CONN_MAX_LIFETIME_S", 300),
			UsePGRepos:       getEnvOrDefault("DB_USE_PG_REPOS", "false") == "true",
		},
		Redis: RedisConfig{
			URL: getEnvOrDefault("REDIS_URL", ""),
		},
		NATS: NATSConfig{
			URL: getEnvOrDefault("NATS_URL", ""),
		},
        Observability: ObservabilityConfig{
            OTELEndpoint:   getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
            OTELInsecure:   getEnvAsBool("OTEL_INSECURE", false),
            OTELHeaders:    getEnvAsKVMap("OTEL_HEADERS"),
            OTELServiceName: getEnvOrDefault("OTEL_SERVICE_NAME", ""),
            MetricsPort:    getEnvOrDefault("METRICS_PORT", "9090"),
        },
		Auth: AuthConfig{
			JWTPublicKey: getEnvOrDefault("JWT_PUBLIC_KEY", ""),
			JWKSURL:      getEnvOrDefault("JWT_JWKS_URL", ""),
			Audience:     getEnvOrDefault("JWT_AUDIENCE", ""),
			Issuer:       getEnvOrDefault("JWT_ISSUER", ""),
			JWKSCacheTTL: getEnvAsInt("JWT_JWKS_CACHE_TTL", 300),
		},
		Storage: StorageConfig{
			S3: S3Config{
				Endpoint:       getEnvOrDefault("STORAGE_S3_ENDPOINT", ""),
				Bucket:         getEnvOrDefault("STORAGE_S3_BUCKET", ""),
				Prefix:         getEnvOrDefault("STORAGE_S3_PREFIX", "snapshots"),
				AccessKey:      getEnvOrDefault("STORAGE_S3_ACCESS_KEY", ""),
				SecretKey:      getEnvOrDefault("STORAGE_S3_SECRET_KEY", ""),
				UseTLS:         getEnvAsBool("STORAGE_S3_USE_TLS", true),
				IgnoreGlobs:    getEnvAsList("STORAGE_S3_IGNORE_GLOBS", []string{"*.map", "dist/*", "**/*.min.js"}),
				SizeLimitBytes: getEnvAsInt64("STORAGE_S3_SIZE_LIMIT_BYTES", 1<<20), // 1MiB
			},
		},
		TLS: TLSConfig{
			EnableMTLS:   getEnvAsBool("TLS_ENABLE_MTLS", false),
			CertFile:     getEnvOrDefault("TLS_CERT_FILE", ""),
			KeyFile:      getEnvOrDefault("TLS_KEY_FILE", ""),
			ClientCAFile: getEnvOrDefault("TLS_CLIENT_CA_FILE", ""),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSec: getEnvAsInt("RATE_LIMIT_RPS", 20),
			Burst:          getEnvAsInt("RATE_LIMIT_BURST", 40),
		},
        AllowedEnvironments: normalizeEnvs(getEnvAsList("ALLOWED_ENVS", []string{"dev","staging","prod","production"})),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if c.ServicePort != "" {
		port, err := strconv.Atoi(c.ServicePort)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("invalid port: must be between 1 and 65535")
		}
	}

	if err := c.Vault.Validate(); err != nil {
		return err
	}

if err := c.Database.Validate(); err != nil {
		return err
	}

	if err := c.Storage.Validate(); err != nil {
		return err
	}

	if err := c.TLS.Validate(); err != nil {
		return err
	}

	if c.RateLimit.RequestsPerSec < 0 || c.RateLimit.Burst < 0 {
		return fmt.Errorf("rate limit values must be non-negative")
	}

    if len(c.AllowedEnvironments) == 0 {
        return fmt.Errorf("ALLOWED_ENVS must not be empty")
    }

	return nil
}

// Validate validates Vault configuration
func (v *VaultConfig) Validate() error {
	if v.Address == "" {
		return fmt.Errorf("VAULT_ADDR is required")
	}

	// Validate URL format
	parsed, err := url.Parse(v.Address)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("VAULT_ADDR must be a valid HTTP/HTTPS URL")
	}

	if v.RoleID == "" {
		return fmt.Errorf("VAULT_ROLE_ID is required")
	}

	if v.SecretID == "" {
		return fmt.Errorf("VAULT_SECRET_ID is required")
	}

	return nil
}

// Validate validates Database configuration
func (d *DatabaseConfig) Validate() error {
	if d.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if d.MaxOpenConns < 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be non-negative")
	}

	if d.MaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be non-negative")
	}

	if d.ConnMaxLifetimeS < 0 {
		return fmt.Errorf("DB_CONN_MAX_LIFETIME_S must be non-negative")
	}

	return nil
}

// Validate validates Storage configuration (optional)
func (s *StorageConfig) Validate() error {
	// S3 is optional. If Endpoint or Bucket is set, require access credentials.
	if s.S3.Endpoint != "" || s.S3.Bucket != "" {
		if s.S3.Endpoint == "" { return fmt.Errorf("STORAGE_S3_ENDPOINT is required when S3 is configured") }
		if s.S3.Bucket == "" { return fmt.Errorf("STORAGE_S3_BUCKET is required when S3 is configured") }
		if s.S3.AccessKey == "" { return fmt.Errorf("STORAGE_S3_ACCESS_KEY is required when S3 is configured") }
		if s.S3.SecretKey == "" { return fmt.Errorf("STORAGE_S3_SECRET_KEY is required when S3 is configured") }
	}
	return nil
}

// Validate TLS config (optional)
func (t *TLSConfig) Validate() error {
	if !t.EnableMTLS {
		return nil
	}
	if t.CertFile == "" || t.KeyFile == "" || t.ClientCAFile == "" {
		return fmt.Errorf("mTLS enabled requires TLS_CERT_FILE, TLS_KEY_FILE, and TLS_CLIENT_CA_FILE")
	}
	return nil
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns environment variable as int or default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsBool returns environment variable as bool or default
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "1", "true", "TRUE", "True", "yes", "Y", "y":
			return true
		case "0", "false", "FALSE", "False", "no", "N", "n":
			return false
		}
	}
	return defaultValue
}

// getEnvAsInt64 returns environment variable as int64 or default
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			return v
		}
	}
	return defaultValue
}

// getEnvAsList splits a comma-separated list; trims spaces; empty -> default
func getEnvAsList(key string, def []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := []string{}
		for _, p := range strings.Split(value, ",") {
			p = strings.TrimSpace(p)
			if p != "" { parts = append(parts, p) }
		}
		return parts
	}
	return def
}

// normalizeEnvs trims, lowercases, and de-duplicates environment names.
func normalizeEnvs(in []string) []string {
    m := map[string]struct{}{}
    out := make([]string, 0, len(in))
    for _, e := range in {
        e = strings.ToLower(strings.TrimSpace(e))
        if e == "" { continue }
        if _, ok := m[e]; ok { continue }
        m[e] = struct{}{}
        out = append(out, e)
    }
    return out
}

// getEnvAsKVMap parses comma-separated key=value pairs into a map.
// - Splits on commas; trims spaces
// - Splits on first '=' only so values may contain '='
// - Skips malformed pairs; last key wins
func getEnvAsKVMap(key string) map[string]string {
    raw := os.Getenv(key)
    if raw == "" { return map[string]string{} }
    out := make(map[string]string)
    for _, part := range strings.Split(raw, ",") {
        part = strings.TrimSpace(part)
        if part == "" { continue }
        eq := strings.Index(part, "=")
        if eq <= 0 { continue }
        k := strings.TrimSpace(part[:eq])
        v := strings.TrimSpace(part[eq+1:])
        if k == "" || v == "" { continue }
        out[k] = v
    }
    return out
}
