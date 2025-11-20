# DB Guardian Service - Database Migration & Schema Governance

**Technology:** Go 1.24
**Ports:** REST: 7105, gRPC: 50065
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The DB Guardian Service provides automated database migration validation, schema governance, and security analysis for PostgreSQL databases. It validates migrations before deployment, analyzes schema changes for breaking changes and security issues, recommends indexes for performance optimization, and enforces role-based access control policies across database environments.

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

# Database migrations
make migration NAME=add_users_table  # Create new migration
make migrate-up                       # Run pending migrations
make migrate-down                     # Rollback last migration
make migrate-status                   # Check migration status
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=db-guardian-service
bunx turbo run test --filter=db-guardian-service
bunx turbo run build --filter=db-guardian-service
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
db-guardian-service/
├── cmd/
│   └── db-guardian-service/
│       ├── main.go              # Application entry point
│       └── grpc_bootstrap.go    # gRPC server setup
├── internal/
│   ├── analyzer/                # SQL & schema analysis engine
│   ├── auth/                    # JWT/OIDC validation
│   ├── cache/                   # Redis caching layer
│   ├── config/                  # Configuration management
│   ├── database/                # PostgreSQL connection & transactions
│   ├── dto/                     # Request/response DTOs
│   ├── events/                  # NATS event publishers
│   ├── grpc/                    # gRPC server implementation
│   ├── models/                  # Domain models
│   ├── ratelimit/               # Rate limiting middleware
│   ├── repository/              # Data access layer
│   │   ├── connections_repository.go
│   │   ├── migration_audits_repository.go
│   │   ├── index_recommendations_repository.go
│   │   └── role_policies_repository.go
│   ├── secrets/                 # Vault integration
│   ├── server/                  # HTTP server & routes
│   ├── service/                 # Business logic layer
│   └── telemetry/               # OpenTelemetry setup
├── pkg/
│   └── logger/                  # Structured logging utilities
├── dbguardian/
│   └── v1/                      # Generated protobuf code
├── api/                         # Protobuf definitions
├── migrations/                  # SQL schema migrations
│   └── 00001_init_schema.sql
├── Makefile                     # Build & dev commands
├── go.mod                       # Go dependencies
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Migration Analysis:**
- File: `internal/analyzer/` - SQL parsing and validation engine
- Pattern: Analyzes migration files for breaking changes, performance issues, security risks
- Example: Detects DROP COLUMN, missing indexes on foreign keys, weak role permissions

**Database Repositories:**
- File: `internal/repository/*_repository.go` - Data access patterns
- Pattern: Each repository handles one domain entity (connections, migrations, policies)
- Example: `migration_audits_repository.go` tracks all validated migrations

**Cache Layer:**
- File: `internal/cache/redis.go` - Redis client wrapper
- Pattern: Cache analysis results to avoid re-validating identical migrations
- TTL: Migration analysis results cached for 1 hour

**Event Publishing:**
- File: `internal/events/` - NATS event publishers
- Pattern: Publishes events on migration validation, schema changes, policy violations
- Example: `migration.validated`, `schema.breaking_change`, `policy.violation`

### Dependencies

**Core:**
- `lib/pq` v1.10.9 - PostgreSQL driver for database connections
- `redis/go-redis/v9` v9.16.0 - Redis client for caching and rate limiting
- `nats-io/nats.go` v1.47.0 - NATS messaging for event-driven communication
- `hashicorp/vault/api` v1.22.0 - HashiCorp Vault SDK for secrets management
- `google.golang.org/grpc` v1.63.2 - gRPC server framework

**Observability:**
- `go.opentelemetry.io/otel` v1.38.0 - OpenTelemetry SDK for distributed tracing
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` - HTTP instrumentation
- `prometheus/client_golang` v1.17.0 - Prometheus metrics (via transitive deps)

**Testing:**
- `DATA-DOG/go-sqlmock` v1.5.2 - SQL mock for unit tests
- Standard library `testing` package

## Code Organization Patterns

### Request Handlers
✅ **DO:** Use handler -> service -> repository layering
```go
// internal/server/handlers.go
func (h *Handler) ValidateMigration(w http.ResponseWriter, r *http.Request) {
    req := dto.ValidateMigrationRequest{}
    // Parse request
    result, err := h.service.ValidateMigration(ctx, req)
    // Write response
}
```
❌ **DON'T:** Put business logic in handlers or make direct database calls

### Database Access
✅ **DO:** Use repository pattern with transaction support
```go
// internal/repository/migration_audits_repository.go
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, audit *models.MigrationAudit) error {
    // Use transaction for atomic operations
}
```
❌ **DON'T:** Hardcode SQL in service layer; always use repositories

### Error Handling
✅ **DO:** Return wrapped errors with context
```go
if err != nil {
    return fmt.Errorf("failed to validate migration %s: %w", migrationID, err)
}
```
❌ **DON'T:** Swallow errors or return generic error messages without context

### Testing
✅ **DO:** Test business logic with mocks; integration tests with real DB
```go
// internal/repository/migration_audits_repository_test.go
func TestRepository_Create(t *testing.T) {
    mock, db := sqlmock.New()
    defer db.Close()
    // Setup expectations and test
}
```
❌ **DON'T:** Skip testing error paths or only test happy paths

## API Endpoints

### REST API

**Base URL:** `http://localhost:7105`

**Key Endpoints:**
- `GET /healthz` - Health check with dependency status
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/migrations/validate` - Validate migration SQL before deployment
- `GET /api/v1/migrations/{id}/audit` - Get migration validation audit trail
- `POST /api/v1/schema/analyze` - Analyze schema for breaking changes
- `GET /api/v1/indexes/recommendations` - Get index recommendations for slow queries
- `POST /api/v1/policies/enforce` - Enforce role-based access policies
- `GET /api/v1/connections/{id}` - Get database connection metadata

### gRPC API

**Port:** 50065

**Services:**
- `DBGuardianService` - Core migration validation and schema governance
  - `ValidateMigration` - Validate SQL migration
  - `GetMigrationAudit` - Retrieve audit history
  - `AnalyzeSchema` - Detect breaking changes
  - `RecommendIndexes` - Performance optimization suggestions

**Proto Files:** `api/*.proto`

## Database Schema

**Tables:**

**`migration_audits`** - Tracks all migration validation attempts
- Columns: `id`, `migration_name`, `project_id`, `status`, `issues`, `validated_at`, `validated_by`
- Indexes: `idx_project_migration`, `idx_status`
- Purpose: Audit trail for compliance and troubleshooting

**`connections`** - Database connection registry
- Columns: `id`, `name`, `project_id`, `host`, `database`, `encrypted_credentials`, `created_at`
- Indexes: `idx_project_id`
- Purpose: Manage connections to databases under governance

**`index_recommendations`** - Performance optimization suggestions
- Columns: `id`, `table_name`, `column_names`, `reason`, `impact_score`, `status`, `created_at`
- Indexes: `idx_table_status`, `idx_impact_score`
- Purpose: Track index recommendations and their implementation status

**`role_policies`** - RBAC policies for database roles
- Columns: `id`, `role_name`, `allowed_operations`, `denied_tables`, `enforced`, `created_at`
- Indexes: `idx_role_name`, `idx_enforced`
- Purpose: Enforce security policies across database environments

**Migrations:**
- Location: `migrations/`
- Tool: Goose (via Makefile)
- Commands: `make migrate-up`, `make migrate-down`, `make migrate-status`

## Event Handling

**Published Events:**
- `migration.validated` - When migration passes/fails validation
  - Payload: `{migration_id, project_id, status, issues[], validated_at}`
- `schema.breaking_change` - When breaking change detected in migration
  - Payload: `{migration_id, change_type, affected_tables[], severity}`
- `index.recommended` - When new index recommendation created
  - Payload: `{table_name, columns[], reason, impact_score}`
- `policy.violation` - When role policy violated
  - Payload: `{role_name, operation, table, policy_id}`

**Subscribed Events:**
- `project.created` - Initialize governance for new project
- `database.connection.added` - Register new database for monitoring

## Testing Strategy

### Unit Tests
- Location: `*_test.go` files colocated with source
- Coverage: Target >80%
- Mock: Database calls with `go-sqlmock`, external services with interfaces
- Example: `internal/repository/migration_audits_repository_test.go`

### Integration Tests
- Location: `internal/integration/` (if present)
- Setup: Use Testcontainers for PostgreSQL, Redis
- Pattern: Test full request flow from handler -> service -> repository -> database

### Running Tests
```bash
# All tests with coverage
make test

# Specific package
go test -v ./internal/service/...

# With race detector
go test -race ./...

# Coverage report
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

### Environment Variables
```bash
# Service-specific config
DATABASE_URL=postgresql://user:pass@localhost:5432/dbguardian
MIGRATION_MAX_SIZE_MB=10
VALIDATION_TIMEOUT_SECONDS=30
ENABLE_AUTO_INDEX_RECOMMENDATIONS=true
POLICY_ENFORCEMENT_MODE=audit  # audit | enforce

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=db-guardian-service
SERVICE_PORT=7105
GRPC_PORT=50065
ENV=dev
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
VAULT_ROLE_ID=<role-id>
VAULT_SECRET_ID=<secret-id>
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/db-guardian-service/`
- Accessed via: `internal/secrets/vault.go` using Vault SDK
- Keys: Database connection credentials, encryption keys for sensitive data

## Quick Find Commands

### Find Code
```bash
# Find migration validation logic
rg -n "ValidateMigration" services/db-guardian-service/internal/

# Find SQL queries in repositories
rg -n "SELECT|INSERT|UPDATE|DELETE" services/db-guardian-service/internal/repository/

# Find event publishers
rg -n "Publish.*migration\." services/db-guardian-service/internal/events/

# Find gRPC method implementations
rg -n "func.*ValidateMigration|AnalyzeSchema" services/db-guardian-service/internal/grpc/

# Find cache usage
rg -n "cache\.(Get|Set)" services/db-guardian-service/internal/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "db-guardian-service" --glob "docker-compose*.yml" --glob "*.yaml"

# Find gRPC clients
rg -n "dbguardian.*NewClient" services/
```

## Common Gotchas

- **Migration Validation Timeouts:** Large migrations may timeout; increase `VALIDATION_TIMEOUT_SECONDS` for complex schema changes or run validation asynchronously
- **Connection Pool Exhaustion:** Each validated migration may test against target database; ensure `DB_MAX_CONNECTIONS` is sufficient for concurrent validations
- **Cache Invalidation:** Analysis results are cached; clear cache manually if analysis logic changes: `redis-cli DEL db-guardian:analysis:*`
- **Breaking Change False Positives:** Analyzer may flag safe changes as breaking; review recommendations manually before blocking deployments
- **Vault Secrets Refresh:** Vault tokens expire; service automatically refreshes but may fail on restart if Vault unavailable; check `VAULT_ADDR` connectivity

## Related Services

- **projects-service:** Provides project context for migration validation; DB Guardian queries project metadata
- **secrets-env-service:** Manages database credentials; DB Guardian retrieves connection strings securely
- **notification-service:** Sends alerts on policy violations or critical breaking changes detected
- **observability-service:** Aggregates migration audit logs and index recommendation metrics

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-db-guardian-service.yml`
- Migration Design: `PHASE1_COMPLETE.md`, `PHASE2_COMPLETE.md`
