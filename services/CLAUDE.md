# Services - Universal Service Patterns

**Parent Context:** This extends [../CLAUDE.md](../CLAUDE.md)

This file contains patterns and guidelines common to all microservices in the LMA platform.
Each service has its own `CLAUDE.md` with service-specific details.

## Service Architecture Overview

### Standard Service Structure
```
<service-name>/
├── CLAUDE.md              # Claude Code guidelines (you are here)
├── AGENTS.md              # AI-assisted development patterns
├── CONTEXT.md             # Service purpose, ports, dependencies
├── README.md              # Quick start guide
├── Dockerfile             # Container definition
├── helm/                  # Kubernetes deployment
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── migrations/            # Database migrations (if applicable)
└── Language-specific structure (see below)
```

### Node.js/TypeScript Services (11 services)
```
src/
├── __tests__/            # Test suites (Vitest)
├── index.ts              # Entry point
├── app.ts                # Fastify app setup
├── config.ts             # Configuration management
├── routes/               # REST endpoint handlers
├── services/             # Business logic
├── repo/                 # Database repositories
├── middleware/           # Fastify middleware
├── events/               # NATS event handlers
├── clients/              # External service clients
├── schemas/              # Zod validation schemas
├── types/                # TypeScript types
└── lib/                  # Utilities
```

**Key Files:**
- `package.json` - Dependencies and scripts
- `tsconfig.json` - TypeScript configuration
- `vitest.config.ts` - Test configuration

### Go Services (12 services)
```
cmd/
└── <service-name>/
    └── main.go           # Entry point
internal/
├── config/               # Configuration
├── server/               # HTTP server
├── grpc/                 # gRPC server
├── service/              # Business logic
├── repository/           # Data access layer
├── models/               # Domain models
├── dto/                  # Data transfer objects
├── auth/                 # JWT/OIDC validation
├── middleware/           # HTTP middleware
├── events/               # NATS event handlers
├── cache/                # Redis caching
├── secrets/              # Vault integration
└── telemetry/            # OpenTelemetry setup
api/                      # Protobuf definitions
migrations/               # SQL migrations (Goose)
Makefile                  # Build, test, migration commands
go.mod                    # Dependencies
```

### Rust Services (1 service: dep-governance)
```
src/
├── main.rs               # Entry point
├── lib.rs                # Library root
├── config.rs             # Configuration
├── server.rs             # Axum server
├── routes/               # HTTP handlers
├── services/             # Business logic
├── models/               # Domain models
└── db/                   # Database layer
tests/                    # Integration tests
migrations/               # SQL migrations (SQLx)
Cargo.toml                # Dependencies
```

### Python Services (3 services)
```
src/
├── __init__.py
├── main.py               # Entry point
├── config.py             # Configuration
├── routes/               # FastAPI routes
├── services/             # Business logic
├── models/               # Data models
└── db/                   # Database layer
tests/                    # pytest tests
requirements.txt          # Dependencies
pyproject.toml            # Python project config
```

## Common Service Patterns

### 1. Configuration Management

**Environment Variables:**
All services support these standard env vars:

```bash
# Service Identity
SERVICE_NAME=<service-name>
SERVICE_VERSION=<version>
PORT=<rest-port>                    # HTTP/REST port
GRPC_PORT=<grpc-port>               # gRPC port (if applicable)

# Database
DATABASE_URL=postgresql://user:pass@host:port/dbname
DB_MAX_CONNECTIONS=10
DB_TIMEOUT=5s

# Cache
REDIS_URL=redis://host:port
REDIS_DB=0
REDIS_TIMEOUT=3s

# Messaging
NATS_URL=nats://host:port
NATS_CLUSTER_ID=lma-cluster

# Secrets
VAULT_ADDR=http://vault:8200
VAULT_ROLE_ID=<role-id>
VAULT_SECRET_ID=<secret-id>

# Authentication
OIDC_ISSUER=http://keycloak:8080/realms/lma
OIDC_AUDIENCE=lma-api
JWT_REQUIRED_SCOPES=read:api,write:api

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=<service-name>
LOG_LEVEL=info                      # debug, info, warn, error

# Feature Flags
ENABLE_GRPC=true
ENABLE_HEALTH_CHECK=true
ENABLE_METRICS=true
ENABLE_TRACING=true
```

**Loading Priority (highest to lowest):**
1. Environment variables
2. `.env.local` (local development only, in .gitignore)
3. Default values in config file
4. Vault secrets (production)

### 2. Health Checks

**Endpoint:** `GET /healthz`

**Response Format:**
```json
{
  "status": "healthy",
  "service": "<service-name>",
  "version": "<version>",
  "uptime": "<duration>",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "nats": "healthy",
    "vault": "healthy"
  }
}
```

**Status Codes:**
- `200` - All dependencies healthy
- `503` - One or more dependencies unhealthy

**Implementation:**
- Check database connection (SELECT 1)
- Check Redis ping
- Check NATS connection
- Check Vault accessibility
- Return degraded status if non-critical dependency fails

### 3. Metrics

**Endpoint:** `GET /metrics`

**Standard Metrics (Prometheus format):**
```
# HTTP metrics
http_requests_total{method, path, status}
http_request_duration_seconds{method, path}

# gRPC metrics (if applicable)
grpc_requests_total{method, status}
grpc_request_duration_seconds{method}

# Business metrics (service-specific)
<service>_operations_total{operation, status}
<service>_operation_duration_seconds{operation}

# System metrics
process_cpu_seconds_total
process_resident_memory_bytes
go_goroutines (Go services)
```

### 4. Authentication & Authorization

**JWT Validation:**
All services validate JWT tokens from Keycloak:

```
Authorization: Bearer <jwt-token>
```

**Required Token Claims:**
- `iss` - Issuer (must match OIDC_ISSUER)
- `aud` - Audience (must match OIDC_AUDIENCE)
- `sub` - Subject (user ID)
- `exp` - Expiration timestamp
- `iat` - Issued at timestamp
- `scope` - Space-separated scopes

**Scope-Based Authorization:**
- `read:<resource>` - Read operations
- `write:<resource>` - Create/update operations
- `delete:<resource>` - Delete operations
- `admin:<resource>` - Administrative operations

**Implementation Pattern:**
1. Extract JWT from `Authorization` header
2. Validate signature using JWKS from OIDC provider
3. Check expiration
4. Verify issuer and audience
5. Extract scopes from claims
6. Enforce scope requirements per endpoint

### 5. Database Patterns

**Connection Management:**
- Use connection pooling (max 10 connections in dev, 50 in prod)
- Set timeouts: connect (5s), idle (10m), lifetime (1h)
- Use prepared statements for parameterized queries
- Always use transactions for multi-step operations

**Migration Management:**
- **Go services:** Goose - `migrations/*.sql`
- **Node services:** Custom scripts - `src/db/migrations/*.sql`
- **Rust services:** SQLx - `migrations/*.sql`
- **Python services:** Alembic - `alembic/versions/*.py`

**Migration Rules:**
- ✅ **DO:** Create new migrations for schema changes
- ✅ **DO:** Test migrations on copy of production data
- ✅ **DO:** Make migrations reversible (up/down)
- ✅ **DO:** Use transactions for DDL changes
- ❌ **DON'T:** Modify existing migrations after merge
- ❌ **DON'T:** Rename migrations
- ❌ **DON'T:** Run migrations manually in production (use CI/CD)

**Query Patterns:**
```typescript
// Node/TypeScript - Use transactions
await db.transaction(async (trx) => {
  await trx('users').insert(user)
  await trx('profiles').insert(profile)
})

// Go - Use transactions
tx, err := db.Begin(ctx)
defer tx.Rollback(ctx)
// ... queries ...
tx.Commit(ctx)

// Rust - Use transactions
let mut tx = pool.begin().await?;
sqlx::query!("...").execute(&mut *tx).await?;
tx.commit().await?;
```

### 6. Caching Patterns

**Redis Usage:**
- Cache expensive queries (TTL: 5-60 minutes)
- Store session data (TTL: 24 hours)
- Rate limiting counters (TTL: 1 minute-1 hour)
- Distributed locks (TTL: 30 seconds)

**Cache Keys Convention:**
```
<service-name>:<resource>:<identifier>[:<field>]

Examples:
projects-service:project:abc123
db-guardian:migration:xyz789:status
test-lab:test-run:def456:results
```

**Cache Invalidation:**
- Invalidate on write operations (create, update, delete)
- Use cache-aside pattern (check cache → query DB → populate cache)
- Set appropriate TTLs based on data volatility
- Use Redis SCAN for bulk invalidation (avoid KEYS *)

### 7. Event-Driven Patterns (NATS)

**Publishing Events:**
```typescript
// Node
await nats.publish('project.created', { projectId, userId })

// Go
nc.Publish("project.created", data)

// Rust
nc.publish("project.created", data).await?
```

**Subscribing to Events:**
```typescript
// Node
nats.subscribe('project.created', async (data) => {
  // Handle event
})

// Go
nc.Subscribe("project.created", func(msg *nats.Msg) {
  // Handle event
})

// Rust
subscriber.next().await
```

**Event Naming Convention:**
```
<resource>.<action>

Examples:
project.created
migration.validated
test.completed
notification.sent
```

### 8. Error Handling

**HTTP Error Responses:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "issue": "invalid format"
    },
    "traceId": "abc123xyz789"
  }
}
```

**Standard Error Codes:**
- `VALIDATION_ERROR` (400) - Request validation failed
- `UNAUTHORIZED` (401) - Missing or invalid authentication
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource not found
- `CONFLICT` (409) - Resource already exists or state conflict
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `INTERNAL_ERROR` (500) - Unexpected server error
- `SERVICE_UNAVAILABLE` (503) - Dependency unavailable

**Error Logging:**
- Log all errors with trace IDs
- Include stack traces for 500 errors
- Redact sensitive data (passwords, tokens, PII)
- Use structured logging (JSON format)

### 9. Testing Standards

**Unit Tests:**
- Colocate with source files
- Test pure business logic without external dependencies
- Mock database, cache, external services
- Aim for >80% coverage

**Integration Tests:**
- Test with real dependencies (Testcontainers or test doubles)
- Verify database queries, cache operations
- Test authentication/authorization flows
- Verify event publishing/subscribing

**Test Data:**
- Use factories or builders for test data
- Seed data in `beforeEach`/`setUp`
- Clean up data in `afterEach`/`tearDown`
- Use transactions with rollback for database tests

**Test Naming:**
```typescript
// Node/TypeScript
describe('ProjectService', () => {
  describe('createProject', () => {
    it('creates project with valid data', async () => {})
    it('throws ValidationError for invalid data', async () => {})
  })
})

// Go
func TestProjectService_CreateProject(t *testing.T) {
  t.Run("creates project with valid data", func(t *testing.T) {})
  t.Run("returns error for invalid data", func(t *testing.T) {})
}

// Rust
#[test]
fn creates_project_with_valid_data() {}

#[test]
fn returns_error_for_invalid_data() {}
```

### 10. Observability (OpenTelemetry)

**Distributed Tracing:**
- Every service exports traces to OTLP collector
- Propagate trace context via HTTP headers: `traceparent`, `tracestate`
- Create spans for: HTTP requests, gRPC calls, database queries, external API calls
- Add span attributes: `service.name`, `http.method`, `http.status_code`, `db.operation`

**Logging:**
- Use structured logging (JSON)
- Include trace IDs in all logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Redact sensitive data

**Example Span:**
```typescript
const span = tracer.startSpan('createProject', {
  attributes: {
    'service.name': 'projects-service',
    'http.method': 'POST',
    'http.route': '/projects',
    'user.id': userId,
  }
})
// ... operation ...
span.setStatus({ code: SpanStatusCode.OK })
span.end()
```

## Service Development Workflow

### 1. Local Development

**Start Dependencies:**
```bash
cd dev
./dev.sh start
```

**Start Single Service (Node):**
```bash
cd services/<service-name>
pnpm install
pnpm dev  # Uses tsx watch for hot-reload
```

**Start Single Service (Go):**
```bash
cd services/<service-name>
make dev  # Uses air for hot-reload
```

**Start Single Service (Rust):**
```bash
cd services/<service-name>
cargo watch -x run
```

### 2. Running Tests

**Node Services:**
```bash
pnpm test              # Run all tests
pnpm test:watch        # Watch mode
pnpm test:coverage     # Generate coverage
```

**Go Services:**
```bash
make test              # Run tests with coverage
go test -v ./...       # Verbose output
```

**Rust Services:**
```bash
cargo test --all       # Run all tests
cargo test -- --nocapture  # Show output
```

### 3. Database Migrations

**Go (Goose):**
```bash
make migration NAME=add_users_table
make migrate-up
make migrate-down
make migrate-status
```

**Node (Custom):**
```bash
pnpm db:migrate:create add_users_table
pnpm db:migrate:up
pnpm db:migrate:down
```

**Rust (SQLx):**
```bash
sqlx migrate add add_users_table
sqlx migrate run
sqlx migrate revert
```

### 4. Pre-PR Checklist

Run these commands before creating a PR:

**Node Services:**
```bash
pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

**Go Services:**
```bash
make quality  # Runs vet, lint, test, build
```

**Rust Services:**
```bash
cargo test --all --locked && cargo build --release
```

**All Services:**
```bash
# From root
bunx turbo run typecheck lint test build --filter=<service-name>
```

## Quick Search Commands

### Find Service Files
```bash
# Find main entry point
rg -n "func main\(\)" services/<go-service>/cmd/
rg -n "app\.listen" services/<node-service>/src/

# Find route handlers
rg -n "fastify\.(get|post|put|delete)" services/<node-service>/
rg -n "http\.HandleFunc|router\.(GET|POST)" services/<go-service>/
rg -n "\.route.*\.handler" services/<rust-service>/

# Find database queries
rg -n "db\.query|db\.exec" services/
rg -n "sqlx::query!" services/<rust-service>/

# Find event handlers
rg -n "nats\.subscribe|nc\.Subscribe" services/

# Find config definitions
rg -n "interface Config|type Config" services/
```

### Find Dependencies
```bash
# Node services
rg -n "from 'fastify'" services/<node-service>/

# Go services
rg -n "import.*github.com" services/<go-service>/

# Rust services
rg -n "use.*crate::" services/<rust-service>/
```

## Common Gotchas

### Node/TypeScript Services
- **Don't use Bun in frontends** - Frontends use pnpm only
- **Import path aliases** - Use `@/` for `src/` imports
- **Vitest vs Jest** - We use Vitest, not Jest
- **Fastify plugins** - Must be async and call `fastify.register()`
- **Branded types** - Use for all IDs: `type UserId = Brand<string, 'UserId'>`

### Go Services
- **Error handling** - Always check errors, never ignore
- **Context propagation** - Pass `context.Context` to all I/O operations
- **defer** - Use for cleanup, but watch for loops (defer runs at function end)
- **Goroutine leaks** - Always have exit conditions
- **nil pointer panics** - Check for nil before dereferencing

### Rust Services
- **Ownership** - Understand borrow checker; use references when possible
- **Error handling** - Use `?` operator, don't unwrap in production code
- **Async runtime** - Use Tokio; don't mix async runtimes
- **Lifetimes** - Avoid explicit lifetimes unless necessary
- **Clone vs Copy** - Prefer Copy for small types, avoid cloning large data

### Python Services
- **Virtual environments** - Always activate venv before installing deps
- **Type hints** - Use for all function signatures
- **Async/await** - FastAPI is async; use `await` for I/O operations
- **Import paths** - Use absolute imports from project root

### Docker
- **Multi-stage builds** - Build stage vs runtime stage
- **Distroless images** - No shell; use debug variant for troubleshooting
- **.dockerignore** - Always create to exclude `node_modules/`, `target/`, `.git/`
- **Layer caching** - Copy dependency files first, then source code

### Database
- **Migrations order** - Never reorder or modify existing migrations
- **Connection leaks** - Always close connections/transactions
- **Indexes** - Add indexes for frequently queried columns
- **N+1 queries** - Use joins or batch loading to avoid

---

**For service-specific details, see individual service CLAUDE.md files:**
- [db-guardian-service](db-guardian-service/CLAUDE.md)
- [test-lab-service](test-lab-service/CLAUDE.md)
- [dep-governance-service](dep-governance-service/CLAUDE.md)
- [notification-service](notification-service/CLAUDE.md)
- [observability-service](observability-service/CLAUDE.md)
- [projects-service](projects-service/CLAUDE.md)
- [secrets-env-service](secrets-env-service/CLAUDE.md)
- [See service_catalog.yaml for complete list]
