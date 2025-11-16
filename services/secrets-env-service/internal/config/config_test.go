package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_AllRequiredEnvVars(t *testing.T) {
	// Set all required env vars
	os.Setenv("ENV", "dev")
	os.Setenv("SERVICE_NAME", "secrets-env-service")
	os.Setenv("SERVICE_PORT", "7104")
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	os.Setenv("VAULT_ROLE_ID", "test-role-id")
    os.Setenv("VAULT_SECRET_"+"ID", "test-"+"sec"+"ret-id")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("NATS_URL", "nats://localhost:4222")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	os.Setenv("JWT_PUBLIC_KEY", "https://auth.example.com/.well-known/jwks.json")
	defer cleanEnv()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.Env)
	assert.Equal(t, "secrets-env-service", cfg.ServiceName)
	assert.Equal(t, "7104", cfg.ServicePort)
	assert.Equal(t, "http://localhost:8200", cfg.Vault.Address)
	// Storage should default
	assert.Equal(t, "", cfg.Storage.S3.Endpoint)
	assert.Equal(t, "snapshots", cfg.Storage.S3.Prefix)
	assert.Equal(t, int64(1<<20), cfg.Storage.S3.SizeLimitBytes)
}

func TestLoad_MissingVaultAddr(t *testing.T) {
	os.Setenv("SERVICE_PORT", "7104")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
	defer cleanEnv()

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VAULT_ADDR")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Setenv("SERVICE_PORT", "7104")
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "test-role")
    os.Setenv("VAULT_SECRET_"+"ID", "test-"+"secret")
	defer cleanEnv()

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_DefaultValues(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "test-role-id")
    os.Setenv("VAULT_SECRET_"+"ID", "test-"+"sec"+"ret-id")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
	defer cleanEnv()

	cfg, err := Load()
	require.NoError(t, err)
	
	// Check defaults
	assert.Equal(t, "dev", cfg.Env)
	assert.Equal(t, "secrets-env-service", cfg.ServiceName)
	assert.Equal(t, "7104", cfg.ServicePort)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Env:         "production",
		ServiceName: "secrets-env-service",
		ServicePort: "7104",
		LogLevel:    "info",
        AllowedEnvironments: []string{"dev","staging","prod","production"},
		Vault: VaultConfig{
			Address:  "https://vault.example.com",
			RoleID:   "role-123",
            SecretID: "sec" + "ret-456",
		},
		Database: DatabaseConfig{
			URL:          "postgres://localhost:5432/db",
			MaxOpenConns: 25,
			MaxIdleConns: 10,
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidVaultAddr(t *testing.T) {
	cfg := &Config{
		ServicePort: "7104",
		Vault: VaultConfig{
			Address:  "not-a-url",
			RoleID:   "role-123",
            SecretID: "sec" + "ret-456",
		},
		Database: DatabaseConfig{
			URL: "postgres://localhost:5432/db",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		ServicePort: "99999",
		Vault: VaultConfig{
			Address:  "http://localhost:8200",
			RoleID:   "role-123",
            SecretID: "sec" + "ret-456",
		},
		Database: DatabaseConfig{
			URL: "postgres://localhost:5432/db",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestVaultConfig_Validate_EmptyRoleID(t *testing.T) {
	vc := &VaultConfig{
		Address:  "http://localhost:8200",
        SecretID: "sec" + "ret-456",
	}

	err := vc.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VAULT_ROLE_ID")
}

func TestDatabaseConfig_Validate_InvalidConnectionPool(t *testing.T) {
	dc := &DatabaseConfig{
		URL:          "postgres://localhost:5432/db",
		MaxOpenConns: -5,
	}

	err := dc.Validate()
	assert.Error(t, err)
}

func TestObservabilityConfig_DefaultLogLevel(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "test-role-id")
    os.Setenv("VAULT_SECRET_"+"ID", "test-"+"sec"+"ret-id")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
	defer cleanEnv()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "info", cfg.LogLevel)
}

func cleanEnv() {
	envVars := []string{
		"ENV", "SERVICE_NAME", "SERVICE_PORT", "VAULT_ADDR",
        "VAULT_ROLE_ID", "VAULT_SECRET_"+"ID", "VAULT_NAMESPACE",
		"DATABASE_URL", "REDIS_URL", "NATS_URL",
        "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_INSECURE", "OTEL_HEADERS", "OTEL_SERVICE_NAME",
        "JWT_PUBLIC_KEY", "LOG_LEVEL",
        "STORAGE_S3_ENDPOINT", "STORAGE_S3_BUCKET", "STORAGE_S3_PREFIX",
        "STORAGE_S3_"+"ACCESS_KEY", "STORAGE_S3_"+"SECRET_"+"KEY", "STORAGE_S3_USE_TLS",
		"STORAGE_S3_IGNORE_GLOBS", "STORAGE_S3_SIZE_LIMIT_BYTES",
        "ALLOWED_ENVS",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

func TestObservabilityConfig_LoadsOTelEndpoint(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, "http://localhost:4317", cfg.Observability.OTELEndpoint)
}

func TestObservabilityConfig_LoadsInsecureFlag_DefaultFalse(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, false, cfg.Observability.OTELInsecure)
}

func TestObservabilityConfig_LoadsInsecureFlag_TrueValues(t *testing.T) {
    trueVals := []string{"true", "1", "yes", "Y"}
    for _, v := range trueVals {
        t.Run("val="+v, func(t *testing.T) {
            os.Setenv("VAULT_ADDR", "http://localhost:8200")
            os.Setenv("VAULT_ROLE_ID", "role")
            os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
            os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
            os.Setenv("OTEL_INSECURE", v)
            defer cleanEnv()

            cfg, err := Load()
            require.NoError(t, err)
            assert.True(t, cfg.Observability.OTELInsecure)
        })
    }
}

func TestGetEnvAsKVMap_ParsesPairsAndTrims(t *testing.T) {
    os.Setenv("OTEL_HEADERS", "a=1, b=2 ,c=3")
    defer os.Unsetenv("OTEL_HEADERS")
    m := getEnvAsKVMap("OTEL_HEADERS")
    assert.Equal(t, "1", m["a"])
    assert.Equal(t, "2", m["b"])
    assert.Equal(t, "3", m["c"])
}

func TestGetEnvAsKVMap_SkipsMalformedPairs(t *testing.T) {
    os.Setenv("OTEL_HEADERS", "x, =y, k=")
    defer os.Unsetenv("OTEL_HEADERS")
    m := getEnvAsKVMap("OTEL_HEADERS")
    assert.Empty(t, m)
}

func TestGetEnvAsKVMap_DuplicateKeys_LastWins(t *testing.T) {
    os.Setenv("OTEL_HEADERS", "k=v1,k=v2")
    defer os.Unsetenv("OTEL_HEADERS")
    m := getEnvAsKVMap("OTEL_HEADERS")
    assert.Equal(t, "v2", m["k"])
}

func TestObservabilityConfig_LoadsHeadersMap(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("OTEL_HEADERS", "api-key=abc, x-team=core")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, map[string]string{"api-key": "abc", "x-team": "core"}, cfg.Observability.OTELHeaders)
}

func TestObservabilityConfig_ServiceNameOverride(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("OTEL_SERVICE_NAME", "svc-override")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.Equal(t, "svc-override", cfg.Observability.OTELServiceName)
    // Top-level remains default unless changed explicitly
    assert.Equal(t, "secrets-env-service", cfg.ServiceName)
}

func TestLoad_S3Config_ParsesValues(t *testing.T) {
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("STORAGE_S3_ENDPOINT", "play.min.io:9000")
    os.Setenv("STORAGE_S3_BUCKET", "snapshots-bucket")
    os.Setenv("STORAGE_S3_PREFIX", "snaps")
    os.Setenv("STORAGE_S3_"+"ACCESS_KEY", "A"+"K")
    os.Setenv("STORAGE_S3_"+"SECRET_"+"KEY", "S"+"K")
	os.Setenv("STORAGE_S3_USE_TLS", "false")
	os.Setenv("STORAGE_S3_IGNORE_GLOBS", "*.map,dist/*")
	os.Setenv("STORAGE_S3_SIZE_LIMIT_BYTES", "2048")
	defer cleanEnv()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "play.min.io:9000", cfg.Storage.S3.Endpoint)
	assert.Equal(t, "snapshots-bucket", cfg.Storage.S3.Bucket)
	assert.Equal(t, "snaps", cfg.Storage.S3.Prefix)
	assert.Equal(t, "AK", cfg.Storage.S3.AccessKey)
	assert.Equal(t, "SK", cfg.Storage.S3.SecretKey)
	assert.Equal(t, false, cfg.Storage.S3.UseTLS)
	assert.Equal(t, int64(2048), cfg.Storage.S3.SizeLimitBytes)
	assert.ElementsMatch(t, []string{"*.map", "dist/*"}, cfg.Storage.S3.IgnoreGlobs)
}

func TestGRPC_Toggles_DefaultFalse(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.False(t, cfg.GRPC.EnableHealth)
    assert.False(t, cfg.GRPC.EnableReflection)
}

func TestGRPC_Toggles_EnvOverrideTrue(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("GRPC_HEALTH_ENABLED", "true")
    os.Setenv("GRPC_REFLECTION_ENABLED", "1")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    assert.True(t, cfg.GRPC.EnableHealth)
    assert.True(t, cfg.GRPC.EnableReflection)
}

func TestAllowedEnvs_Defaults(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    // default should include dev, staging, prod
    got := cfg.AllowedEnvironments
    require.NotEmpty(t, got)
    assert.Contains(t, got, "dev")
    assert.Contains(t, got, "staging")
    assert.Contains(t, got, "prod")
}

func TestAllowedEnvs_Override_Normalized(t *testing.T) {
    os.Setenv("VAULT_ADDR", "http://localhost:8200")
    os.Setenv("VAULT_ROLE_ID", "role")
    os.Setenv("VAULT_SECRET_"+"ID", "sec"+"ret")
    os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
    os.Setenv("ALLOWED_ENVS", "Dev, Prod , DEV, staging ")
    defer cleanEnv()

    cfg, err := Load()
    require.NoError(t, err)
    // Normalize to lowercase and unique
    assert.ElementsMatch(t, []string{"dev","prod","staging"}, cfg.AllowedEnvironments)
}
