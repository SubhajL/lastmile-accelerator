package config

import (
    "testing"
)

func TestConfig_AuthVarsRequireCoherentSet(t *testing.T) {
    t.Setenv("SERVICE_NAME", "db-guardian-service")
    t.Setenv("SERVICE_PORT", "7105")

    // Only JWKS URL set -> error
    t.Setenv("AUTH_JWKS_URL", "https://auth.example.com/.well-known/jwks.json")
    if _, err := Load(); err == nil {
        t.Fatalf("expected error when only AUTH_JWKS_URL is set")
    }

    // All set coherently -> ok
    t.Setenv("AUTH_ISSUER", "https://auth.example.com/")
    t.Setenv("AUTH_AUDIENCE", "db-guardian")
    t.Setenv("AUTH_CLOCK_SKEW_SECONDS", "60")
    if _, err := Load(); err != nil {
        t.Fatalf("unexpected error with coherent auth config: %v", err)
    }
}

func TestConfig_AuthVars_InvalidURLAndSkew(t *testing.T) {
    t.Setenv("SERVICE_NAME", "db-guardian-service")
    t.Setenv("SERVICE_PORT", "7105")

    // Invalid URL -> error
    t.Setenv("AUTH_JWKS_URL", "not-a-url")
    t.Setenv("AUTH_ISSUER", "https://issuer/")
    t.Setenv("AUTH_AUDIENCE", "aud")
    t.Setenv("AUTH_CLOCK_SKEW_SECONDS", "10")
    if _, err := Load(); err == nil {
        t.Fatalf("expected error for invalid AUTH_JWKS_URL")
    }

    // Negative skew -> error
    t.Setenv("AUTH_JWKS_URL", "https://issuer/.well-known/jwks.json")
    t.Setenv("AUTH_CLOCK_SKEW_SECONDS", "-5")
    if _, err := Load(); err == nil {
        t.Fatalf("expected error for negative AUTH_CLOCK_SKEW_SECONDS")
    }
}
