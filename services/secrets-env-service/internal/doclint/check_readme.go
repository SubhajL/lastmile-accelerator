package doclint

import (
    "bytes"
    "errors"
    "fmt"
    "regexp"
    "strings"
)

func ValidateSections(data []byte) error {
    want := []string{
        "## Overview",
        "## Quickstart",
        "## Configuration",
        "## Endpoints",
        "## Readiness and Health",
        "## Metrics and Observability",
        "## Security and RBAC",
        "## gRPC",
        "## Development",
        "## Troubleshooting",
    }
    s := string(data)
    last := -1
    for _, h := range want {
        i := strings.Index(s, h)
        if i < 0 { return fmt.Errorf("missing heading: %s", h) }
        if i < last { return errors.New("headings out of order") }
        last = i
    }
    return nil
}

func ValidateEnvKeys(data []byte) error {
    missing := []string{}
    for _, k := range requiredEnvKeys() {
        if !bytes.Contains(bytes.ToUpper(data), []byte(strings.ToUpper(k))) {
            missing = append(missing, k)
        }
    }
    if len(missing) > 0 { return fmt.Errorf("env keys missing: %v", missing) }
    return nil
}

func ValidateEndpoints(data []byte) error {
    for _, e := range requiredEndpoints() {
        if !bytes.Contains(data, []byte(e)) { return fmt.Errorf("endpoint missing: %s", e) }
    }
    return nil
}

func ValidateObservability(data []byte) error {
    need := []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_INSECURE", "OTEL_HEADERS", "OTEL_SERVICE_NAME", "/metrics"}
    for _, s := range need { if !bytes.Contains(data, []byte(s)) { return fmt.Errorf("observability missing: %s", s) } }
    return nil
}

func ValidateGRPCToggles(data []byte) error {
    need := []string{"GRPC_HEALTH_ENABLED", "GRPC_REFLECTION_ENABLED", "health", "reflection"}
    for _, s := range need { if !bytes.Contains(bytes.ToLower(data), []byte(strings.ToLower(s))) { return fmt.Errorf("grpc toggle mention missing: %s", s) } }
    return nil
}

func ValidateRBAC(data []byte) error {
    s := strings.ToLower(string(data))
    if !strings.Contains(s, "admin") || !strings.Contains(s, "auditor") { return errors.New("roles admin/auditor not documented") }
    if !strings.Contains(s, "secrets:read") { return errors.New("scope secrets:read not documented") }
    return nil
}

func ValidateQuickstart(data []byte) error {
    s := string(data)
    if !strings.Contains(s, "make build") || !strings.Contains(s, "make test") { return errors.New("quickstart missing build/test commands") }
    return nil
}

func ValidateTroubleshooting(data []byte) error {
    s := strings.ToLower(string(data))
    patterns := []*regexp.Regexp{
        regexp.MustCompile(`otlp|otel`),
        regexp.MustCompile(`vault|database|postgres`),
    }
    for _, rx := range patterns {
        if rx.FindStringIndex(s) == nil { return fmt.Errorf("troubleshooting missing topic: %s", rx.String()) }
    }
    return nil
}

func requiredEnvKeys() []string {
    return []string{
        "ENV", "SERVICE_NAME", "SERVICE_PORT", "LOG_LEVEL",
        "VAULT_ADDR", "VAULT_ROLE_ID", "VAULT_SECRET_ID",
        "DATABASE_URL", "REDIS_URL", "NATS_URL",
        "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_INSECURE", "OTEL_HEADERS", "OTEL_SERVICE_NAME",
        "JWT_PUBLIC_KEY",
        "STORAGE_S3_ENDPOINT", "STORAGE_S3_BUCKET", "STORAGE_S3_PREFIX",
        "STORAGE_S3_ACCESS_KEY", "STORAGE_S3_SECRET_KEY", "STORAGE_S3_USE_TLS",
        "STORAGE_S3_IGNORE_GLOBS", "STORAGE_S3_SIZE_LIMIT_BYTES",
        "ALLOWED_ENVS", "HTTP_MAX_BODY_BYTES",
        "GRPC_HEALTH_ENABLED", "GRPC_REFLECTION_ENABLED",
        "STARTUP_CRITICAL_TIMEOUT_S", "STARTUP_OPTIONAL_TIMEOUT_S",
    }
}

func requiredEndpoints() []string {
    return []string{
        "/healthz", "/readyz", "/metrics",
        "/v1/projects/{projectID}/secrets",
        "/v1/projects/{projectID}/secrets/{key}",
        "/v1/projects/{projectID}/env-parity",
        "/v1/projects/{projectID}/env-parity/latest",
        "/v1/projects/{projectID}/env-parity/history",
        "/v1/projects/{projectID}/scan/client-leaks",
    }
}
