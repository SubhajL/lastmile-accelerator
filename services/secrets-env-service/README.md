# secrets-env-service

## Overview
Service for environment-scoped secrets management with drift parity checks and leak-scan intake. Provides HTTP and gRPC APIs, metrics, OTEL traces, and readiness/health endpoints.

## Quickstart
```bash
make build && make test && make run
```

## Configuration
Environment variables (key subset shown; defaults come from code):

| Key | Default | Description |
| --- | --- | --- |
| ENV | dev | Deployment environment name |
| SERVICE_NAME | secrets-env-service | Service identifier |
| SERVICE_PORT | 7104 | HTTP listen port |
| LOG_LEVEL | info | Log level |
| VAULT_ADDR | (required) | Vault address |
| VAULT_ROLE_ID | (required) | AppRole role id |
| VAULT_SECRET_ID | (required) | AppRole secret id |
| DATABASE_URL | (required) | Postgres DSN when PG repos enabled |
| REDIS_URL |  | Optional Redis cache URL |
| NATS_URL |  | Optional NATS URL for events |
| OTEL_EXPORTER_OTLP_ENDPOINT |  | OTLP endpoint (http(s)://host:port) |
| OTEL_INSECURE | false | Allow insecure OTLP transport |
| OTEL_HEADERS |  | Comma key=value list, sent as headers |
| OTEL_SERVICE_NAME |  | Override OTEL service name |
| JWT_PUBLIC_KEY |  | JWKS URL for JWT verification |
| STORAGE_S3_ENDPOINT |  | S3 endpoint (minio compatible) |
| STORAGE_S3_BUCKET |  | S3 bucket |
| STORAGE_S3_PREFIX | snapshots | S3 key prefix |
| STORAGE_S3_ACCESS_KEY |  | Access key |
| STORAGE_S3_SECRET_KEY |  | Secret key |
| STORAGE_S3_USE_TLS | true | Use TLS to S3 endpoint |
| STORAGE_S3_IGNORE_GLOBS |  | Ignore patterns (comma-separated) |
| STORAGE_S3_SIZE_LIMIT_BYTES | 1048576 | Upload size limit |
| ALLOWED_ENVS | dev,staging,prod,production | Allowed env names |
| HTTP_MAX_BODY_BYTES | 1048576 | Request body limit |
| GRPC_HEALTH_ENABLED | false | Enable gRPC health service |
| GRPC_REFLECTION_ENABLED | false | Enable gRPC reflection service |
| STARTUP_CRITICAL_TIMEOUT_S | 3 | Timeout for critical checks |
| STARTUP_OPTIONAL_TIMEOUT_S | 1 | Timeout for optional checks |

## Endpoints
- /healthz — basic process health
- /readyz — readiness (PG/Vault must be healthy)
- /metrics — Prometheus metrics
- /v1/projects/{projectID}/secrets (POST, GET)
- /v1/projects/{projectID}/secrets/{key} (GET, DELETE)
- /v1/projects/{projectID}/env-parity (POST)
- /v1/projects/{projectID}/env-parity/latest (GET)
- /v1/projects/{projectID}/env-parity/history (GET)
- /v1/projects/{projectID}/scan/client-leaks (POST)
- /v1/projects/{projectID}/scan/client-leaks/{snapshotID} (GET)
- /v1/projects/{projectID}/scan/client-leaks/{scanID}/fix (PATCH)

## Readiness and Health
- /readyz returns 200 when startup checks pass, else 503. Critical: Postgres (if enabled) and Vault. Optional: Redis/NATS/S3 do not block readiness but are reported.

## Metrics and Observability
- /metrics exposes counters/histograms for HTTP and gRPC.
- OTEL tracing via OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_INSECURE, OTEL_HEADERS, and optional OTEL_SERVICE_NAME override.

## Security and RBAC
- JWT verification uses JWKS (JWT_PUBLIC_KEY). Roles from JWT claims map to scopes:
  - admin → secrets:read, secrets:write, parity:read, parity:compute, leaks:read, leaks:write
  - auditor → secrets:read, parity:read, leaks:read

## gRPC
- Serves SecretsEnvService; enable health and reflection via GRPC_HEALTH_ENABLED and GRPC_REFLECTION_ENABLED.

## Development
```bash
make test
make build
```

## Troubleshooting
- OTLP exporter failing: verify OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_INSECURE.
- Vault or Postgres unavailable: /readyz returns 503 until healthy.
