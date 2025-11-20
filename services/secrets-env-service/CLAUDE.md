# Secrets & Environment Service - Centralized Secrets Management

**Technology:** Go 1.24
**Ports:** REST: 7104, gRPC: 50064
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Secrets & Environment Service provides centralized secrets management and environment variable configuration for the LMA platform. It integrates with HashiCorp Vault for secure secret storage, manages environment-specific configurations (dev, staging, prod), provides dynamic secret rotation and lease management, enforces access control policies for secret retrieval, and offers unified APIs for services to fetch secrets without direct Vault interaction. This service abstracts Vault complexity and provides a single source of truth for all platform secrets.

## Development Commands

### From This Directory
```bash
# Go service commands
go mod download        # Download dependencies
make dev              # Hot-reload with air
make test             # Run tests with coverage
make build            # Build binary
make lint             # Run golangci-lint
make quality          # Run all quality checks (vet, lint, test, build)

# Database migrations (if applicable)
make migration NAME=add_secret_metadata
make migrate-up
make migrate-down
make migrate-status
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=secrets-env-service
bunx turbo run test --filter=secrets-env-service
bunx turbo run build --filter=secrets-env-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
make quality

# Or individually
go vet ./...
golangci-lint run
go test -race -cover ./...
go build ./...
```

## Architecture

### Directory Structure
```
secrets-env-service/
├── cmd/
│   └── secrets-env-service/
│       └── main.go              # Application entry point
├── internal/
│   ├── cache/                   # Redis caching for secrets
│   │   ├── redis.go
│   │   └── redis_test.go
│   ├── config/                  # Configuration management
│   ├── domain/                  # Domain models
│   │   ├── secret.go
│   │   └── lease.go
│   ├── errors/                  # Custom error types
│   ├── events/                  # NATS event publishing
│   ├── grpc/                    # gRPC server implementation
│   ├── handlers/                # HTTP request handlers
│   │   ├── secrets.go
│   │   ├── env.go
│   │   └── health.go
│   ├── logger/                  # Structured logging (zerolog)
│   ├── observability/           # OpenTelemetry setup
│   ├── repository/              # Secret metadata storage
│   │   ├── secrets.repo.go
│   │   └── leases.repo.go
│   ├── security/                # Encryption & validation
│   │   ├── encryption.go
│   │   └── policy.go
│   ├── server/                  # HTTP server (chi router)
│   ├── service/                 # Business logic
│   │   ├── secrets.service.go
│   │   ├── vault.service.go
│   │   └── rotation.service.go
│   └── vault/                   # Vault client wrapper
│       ├── client.go
│       └── auth.go
├── db/                          # Database migrations (if applicable)
│   └── migrations/
├── packages/                    # Shared packages
├── go.mod                       # Go dependencies
├── Makefile                     # Build & dev commands
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Vault Client:**
- File: `internal/vault/client.go` - HashiCorp Vault SDK wrapper
- Pattern: Authenticates with AppRole, manages token renewal, handles retries
- Example: Fetches secret from `secret/data/projects-service/database` path

**Secrets Service:**
- File: `internal/service/secrets.service.go` - Secret retrieval and management
- Pattern: Checks cache, fetches from Vault, stores in cache with TTL
- Example: GetSecret() returns cached secret or fetches from Vault

**Rotation Service:**
- File: `internal/service/rotation.service.go` - Automatic secret rotation
- Pattern: Monitors lease expiration, rotates credentials, publishes rotation events
- Example: Rotates database passwords before lease expiration

**Cache Layer:**
- File: `internal/cache/redis.go` - Redis caching for frequently accessed secrets
- Pattern: Cache secrets with short TTL (5-15 minutes) to reduce Vault load
- Example: Caches database connection strings to avoid repeated Vault requests

**Access Policy Enforcer:**
- File: `internal/security/policy.go` - Validates service permissions
- Pattern: Checks if requesting service has access to requested secret path
- Example: Blocks test-lab-service from accessing production secrets

### Dependencies

**Core:**
- `hashicorp/vault/api` v1.22.0 - HashiCorp Vault SDK for secret management
- `jackc/pgx/v5` v5.7.6 - PostgreSQL driver for metadata storage
- `redis/go-redis/v9` v9.16.0 - Redis for caching secrets
- `nats-io/nats.go` v1.47.0 - NATS messaging for rotation events
- `go-chi/chi/v5` v5.2.3 - Lightweight HTTP router
- `golang-jwt/jwt/v5` v5.3.0 - JWT validation for API access
- `google.golang.org/grpc` v1.76.0 - gRPC server for high-performance access

**Security:**
- `rs/zerolog` v1.34.0 - Structured logging with redaction support
- `google.golang.org/protobuf` v1.36.8 - Protobuf for gRPC

**Storage:**
- `minio/minio-go/v7` v7.0.97 - S3-compatible storage for secret backups

**Observability:**
- `go.opentelemetry.io/otel` v1.37.0 - OpenTelemetry SDK
- `go.opentelemetry.io/otel/sdk` v1.37.0 - OTel SDK
- `go.opentelemetry.io/otel/trace` v1.37.0 - Distributed tracing
- `prometheus/client_golang` v1.23.2 - Prometheus metrics

**Testing:**
- `DATA-DOG/go-sqlmock` v1.5.2 - SQL mocking
- `alicebob/miniredis/v2` v2.35.0 - Redis mock for tests
- `stretchr/testify` v1.11.1 - Test assertions

## Code Organization Patterns

### Secret Retrieval
✅ **DO:** Use cache-aside pattern for secret access
```go
// internal/service/secrets.service.go
func (s *Service) GetSecret(ctx context.Context, path string) (*Secret, error) {
    // Check cache first
    if cached, err := s.cache.Get(ctx, path); err == nil {
        return cached, nil
    }
    // Fetch from Vault
    secret, err := s.vault.Read(ctx, path)
    if err != nil {
        return nil, err
    }
    // Cache for 10 minutes
    s.cache.Set(ctx, path, secret, 10*time.Minute)
    return secret, nil
}
```
❌ **DON'T:** Always hit Vault; use caching to reduce latency and load

### Token Renewal
✅ **DO:** Automatically renew Vault tokens before expiration
```go
// internal/vault/auth.go
func (c *Client) startTokenRenewal(ctx context.Context) {
    ticker := time.NewTicker(c.tokenTTL / 2)
    go func() {
        for range ticker.C {
            c.renewToken(ctx)
        }
    }()
}
```
❌ **DON'T:** Wait for token expiration; renew proactively

### Access Control
✅ **DO:** Validate requesting service against ACL policies
```go
// internal/security/policy.go
func (p *Policy) CanAccess(service, secretPath string) bool {
    return p.acl[service].Contains(secretPath)
}
```
❌ **DON'T:** Allow unrestricted secret access; enforce least privilege

### Error Handling
✅ **DO:** Redact secret values from error messages and logs
```go
// internal/logger/logger.go
func (l *Logger) Error(msg string, err error, fields ...Field) {
    // Redact sensitive fields before logging
    safeFields := redactSecrets(fields)
    l.log.Error().Err(err).Fields(safeFields).Msg(msg)
}
```
❌ **DON'T:** Log secrets or include them in error responses

## API Endpoints

### REST API

**Base URL:** `http://localhost:7104`

**Key Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/secrets/{path}` - Retrieve secret by path
- `POST /api/v1/secrets` - Create/update secret in Vault
- `DELETE /api/v1/secrets/{path}` - Delete secret
- `GET /api/v1/env/{service}/{environment}` - Get environment variables for service
- `POST /api/v1/rotate/{path}` - Trigger manual secret rotation
- `GET /api/v1/leases/{id}` - Get lease information
- `POST /api/v1/leases/{id}/renew` - Renew lease

### gRPC API

**Port:** 50064

**Services:**
- `SecretsService` - High-performance secret retrieval
  - `GetSecret` - Fetch secret by path
  - `GetEnvVars` - Get environment variables
  - `RotateSecret` - Trigger rotation
  - `RenewLease` - Extend lease

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`secret_metadata`** - Secret access audit trail
- Columns: `id`, `path`, `accessed_by`, `accessed_at`, `operation` (read|write|rotate|delete)
- Indexes: `idx_path`, `idx_accessed_by`, `idx_accessed_at`
- Purpose: Audit all secret access for compliance

**`leases`** - Active secret leases
- Columns: `lease_id`, `secret_path`, `service`, `expires_at`, `renewable`, `created_at`
- Indexes: `idx_lease_id`, `idx_expires_at`, `idx_secret_path`
- Purpose: Track dynamic secret leases and expiration

**`rotation_schedules`** - Automatic rotation configuration
- Columns: `id`, `secret_path`, `rotation_interval`, `last_rotated`, `next_rotation`
- Indexes: `idx_next_rotation`, `idx_secret_path`
- Purpose: Schedule automatic credential rotation

**Migrations:**
- Location: `db/migrations/`
- Tool: Goose or custom migration framework
- Commands: `make migrate-up`, `make migrate-down`

## Event Handling

**Published Events:**
- `secret.rotated` - When secret rotated
  - Payload: `{path, rotated_at, rotated_by, new_lease_id}`
- `secret.accessed` - When secret retrieved (audit)
  - Payload: `{path, accessed_by, accessed_at, cache_hit}`
- `lease.expiring` - When lease nearing expiration
  - Payload: `{lease_id, secret_path, expires_at, time_remaining}`
- `secret.created` - When new secret stored
  - Payload: `{path, created_by, created_at}`

**Subscribed Events:**
- `service.deployed` - Rotate secrets for newly deployed service
- `user.offboarded` - Revoke user's personal secrets

## Testing Strategy

### Unit Tests
- Location: `*_test.go` files colocated with source
- Coverage: Target >80%
- Mock: Vault with in-memory implementation, Redis with miniredis, PostgreSQL with sqlmock
- Example: `internal/vault/client_test.go`

### Integration Tests
- Location: `internal/integration/` (if present)
- Setup: Use Testcontainers for Vault, PostgreSQL, Redis
- Pattern: Test full secret lifecycle (create -> read -> rotate -> delete)

### Running Tests
```bash
# All tests with coverage
make test

# Specific package
go test -v ./internal/vault/...

# With race detector
go test -race ./...

# Coverage report
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

### Environment Variables
```bash
# Vault configuration
VAULT_ADDR=http://vault:8200
VAULT_ROLE_ID=<approle-role-id>
VAULT_SECRET_ID=<approle-secret-id>
VAULT_TOKEN_RENEW_INTERVAL=1h
VAULT_MAX_RETRIES=3

# Service-specific config
SECRET_CACHE_TTL=10m
ENABLE_SECRET_ROTATION=true
ROTATION_CHECK_INTERVAL=5m
MAX_LEASE_DURATION=24h

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=secrets-env-service
SERVICE_PORT=7104
GRPC_PORT=50064
ENV=dev
DATABASE_URL=postgresql://user:pass@localhost:5432/secrets
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/secrets-env-service/`
- Accessed via: Self-referential (uses own Vault integration)
- Keys: Vault AppRole credentials, database encryption keys

## Quick Find Commands

### Find Code
```bash
# Find Vault client usage
rg -n "vault\.Read|vault\.Write" services/secrets-env-service/internal/

# Find secret retrieval logic
rg -n "GetSecret|FetchSecret" services/secrets-env-service/internal/

# Find rotation logic
rg -n "RotateSecret|rotation" services/secrets-env-service/internal/

# Find cache usage
rg -n "cache\.(Get|Set)" services/secrets-env-service/internal/

# Find access control
rg -n "CanAccess|policy.*check" services/secrets-env-service/internal/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "secrets-env-service|localhost:7104" --glob "docker-compose*.yml" --glob "*.yaml"

# Find services retrieving secrets
rg -n "getSecret|fetchSecret|secretsClient" services/
```

## Common Gotchas

- **Vault Token Expiration:** Vault tokens expire; implement automatic renewal or service crashes when token invalid
- **Cache Invalidation:** Rotated secrets may be cached; ensure cache eviction on rotation or use short TTLs
- **Lease Management:** Dynamic secrets have leases; track and renew leases before expiration or credentials become invalid
- **Secret Redaction:** Secrets may leak in logs or error messages; always redact sensitive values before logging
- **Vault Unavailability:** Service degrades when Vault unavailable; implement circuit breaker and fallback to cached secrets
- **Rate Limiting:** Vault has rate limits; implement request batching and exponential backoff for retries
- **Encryption Keys:** Encryption keys for local storage must be rotated; use Vault Transit engine or KMS
- **Concurrent Access:** Concurrent secret rotations may cause conflicts; use distributed locks (Redis) to coordinate

## Related Services

- **All Services:** Every service depends on Secrets Service for credentials and configuration
- **db-guardian-service:** Retrieves database connection secrets for migration validation
- **projects-service:** Fetches project-specific secrets and API keys
- **notification-service:** Accesses SMTP/Twilio credentials for sending notifications
- **observability-service:** Uses secrets for external monitoring service integrations

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-secrets-env-service.yml`
- HashiCorp Vault Docs: https://www.vaultproject.io/docs
- Vault AppRole Auth: https://www.vaultproject.io/docs/auth/approle
- Best Practices: `TASKS_1-6_COMPLETED.md`, `TASKS_7-9_COMPLETE.md`
