# Dependency Governance Service - SBOM & Vulnerability Management

**Technology:** Rust 2021 Edition, Axum
**Ports:** REST: 7106, gRPC: 50066
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Dependency Governance Service manages Software Bill of Materials (SBOM) generation, dependency tracking, and vulnerability scanning for all projects in the LMA platform. It automatically generates and validates SBOMs in CycloneDX/SPDX formats, scans dependencies for known vulnerabilities using CVE databases, enforces dependency policies (license compliance, version pinning), and provides remediation recommendations for security issues.

## Development Commands

### From This Directory
```bash
# Rust service commands
cargo build                # Build in debug mode
cargo build --release      # Build optimized binary
cargo watch -x run         # Hot-reload during development
cargo test --all           # Run all tests
cargo test --all --locked  # Run tests without updating dependencies
cargo test -- --nocapture  # Run tests with output
cargo clippy               # Lint with Clippy
cargo clippy -- -D warnings  # Lint with warnings as errors
cargo fmt                  # Format code
cargo check                # Fast compilation check

# Database migrations
sqlx migrate add create_sboms_table
sqlx migrate run
sqlx migrate revert
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=dep-governance-service
bunx turbo run test --filter=dep-governance-service
bunx turbo run build --filter=dep-governance-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
cargo fmt --check && cargo clippy -- -D warnings && cargo test --all --locked && cargo build --release

# Or use Makefile
make quality
```

## Architecture

### Directory Structure
```
dep-governance-service/
├── src/
│   ├── main.rs                  # Application entry point
│   ├── lib.rs                   # Library root
│   ├── config/                  # Configuration management
│   │   ├── mod.rs
│   │   └── app.rs              # AppConfig struct
│   ├── db/                      # Database layer
│   │   ├── mod.rs
│   │   ├── migrate.rs          # Migration runner
│   │   ├── sboms.rs            # SBOM repository
│   │   ├── dependencies.rs     # Dependency repository
│   │   └── vulnerabilities.rs  # Vulnerability repository
│   ├── models/                  # Domain models
│   │   ├── mod.rs
│   │   ├── sbom.rs             # SBOM types
│   │   ├── dependency.rs       # Dependency types
│   │   ├── vulnerability.rs    # CVE types
│   │   └── common.rs           # Shared types
│   ├── handlers/                # HTTP request handlers
│   │   ├── mod.rs
│   │   ├── sbom.rs             # SBOM CRUD endpoints
│   │   ├── scan.rs             # Vulnerability scanning
│   │   ├── healthz.rs          # Health check
│   │   ├── readyz.rs           # Readiness check
│   │   └── metrics.rs          # Prometheus metrics
│   ├── services/                # Business logic
│   │   ├── sbom_generator.rs   # SBOM generation
│   │   ├── scanner.rs          # Vulnerability scanning
│   │   └── policy_engine.rs    # Policy enforcement
│   ├── middleware/              # Axum middleware
│   │   ├── mod.rs
│   │   ├── auth.rs             # JWT validation
│   │   └── telemetry.rs        # Tracing middleware
│   ├── events/                  # NATS event handling
│   │   ├── publisher.rs
│   │   └── subscriber.rs
│   └── error.rs                 # Error types
├── tests/                       # Integration tests
│   ├── integration_test.rs
│   ├── sbom_tests.rs
│   └── scan_tests.rs
├── migrations/                  # SQL migrations (SQLx)
│   ├── 001_create_sboms.sql
│   ├── 002_create_dependencies.sql
│   └── 003_create_vulnerabilities.sql
├── Cargo.toml                   # Dependencies and metadata
├── Cargo.lock                   # Locked dependency versions
├── .env.example                 # Example environment variables
├── Makefile                     # Build shortcuts
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**SBOM Generation:**
- File: `src/services/sbom_generator.rs` - Generates SBOMs from package manifests
- Pattern: Parses `package.json`, `Cargo.toml`, `go.mod`, `pom.xml` and generates CycloneDX/SPDX
- Example: Converts `package.json` dependencies to CycloneDX JSON format

**Vulnerability Scanner:**
- File: `src/services/scanner.rs` - Scans dependencies against CVE databases
- Pattern: Queries OSV, NVD, GitHub Advisory Database for known vulnerabilities
- Example: Detects CVE-2023-12345 in lodash@4.17.20

**Policy Engine:**
- File: `src/services/policy_engine.rs` - Enforces dependency policies
- Pattern: Validates licenses, version constraints, security policies
- Example: Blocks GPL dependencies in proprietary projects

**Database Repositories:**
- File: `src/db/sboms.rs`, `src/db/dependencies.rs`, `src/db/vulnerabilities.rs`
- Pattern: SQLx-based async database access with compile-time query verification
- Example: Uses `sqlx::query!` macro for type-safe queries

**Event Publishing:**
- File: `src/events/publisher.rs` - NATS event publishing
- Pattern: Publishes events on SBOM generation, vulnerabilities detected
- Example: `sbom.generated`, `vulnerability.detected`

### Dependencies

**Core:**
- `axum` 0.7 - Web framework built on Tokio and Hyper
- `tower` 0.4 - Middleware and service abstractions
- `tower-http` 0.5 - HTTP middleware (tracing, CORS)
- `tokio` 1.35 - Async runtime with full features
- `sqlx` 0.7 - Async PostgreSQL driver with migrations
- `redis` 0.24 - Redis client with Tokio integration
- `async-nats` 0.33 - NATS messaging client

**Serialization:**
- `serde` 1.0 - Serialization framework
- `serde_json` 1.0 - JSON support

**Security & Auth:**
- `jsonwebtoken` 9.2 - JWT validation for authentication

**Observability:**
- `tracing` 0.1 - Structured logging and tracing
- `tracing-subscriber` 0.3 - Tracing collector with JSON output
- `opentelemetry` 0.21 - OpenTelemetry SDK
- `opentelemetry-otlp` 0.14 - OTLP exporter for traces
- `tracing-opentelemetry` 0.22 - Bridge between tracing and OTel

**Utilities:**
- `uuid` 1.6 - UUID generation with v4 support
- `chrono` 0.4 - Date and time handling
- `anyhow` 1.0 - Flexible error handling
- `thiserror` 1.0 - Derive macros for custom errors
- `config` 0.13 - Configuration management
- `dotenvy` 0.15 - .env file loading

**Testing:**
- `axum-test` 14.0 - HTTP testing for Axum
- `testcontainers` 0.15 - Docker containers for integration tests
- `wiremock` 0.6 - HTTP mocking

## Code Organization Patterns

### Request Handlers
✅ **DO:** Use handler -> service -> repository pattern
```rust
// src/handlers/sbom.rs
pub async fn create_sbom(
    State(state): State<AppState>,
    Json(req): Json<CreateSbomRequest>,
) -> Result<Json<Sbom>, AppError> {
    let sbom = state.sbom_service.generate(req).await?;
    Ok(Json(sbom))
}
```
❌ **DON'T:** Put business logic in handlers or make direct database calls

### Database Access
✅ **DO:** Use SQLx with compile-time verified queries
```rust
// src/db/sboms.rs
pub async fn create(pool: &PgPool, sbom: &Sbom) -> Result<Uuid, sqlx::Error> {
    let id = sqlx::query!(
        "INSERT INTO sboms (project_id, format, content) VALUES ($1, $2, $3) RETURNING id",
        sbom.project_id,
        sbom.format,
        sbom.content
    )
    .fetch_one(pool)
    .await?
    .id;
    Ok(id)
}
```
❌ **DON'T:** Use string concatenation for SQL; always use parameterized queries

### Error Handling
✅ **DO:** Use `thiserror` for custom errors and `anyhow` for internal errors
```rust
// src/error.rs
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("Vulnerability not found: {0}")]
    VulnerabilityNotFound(String),
}
```
❌ **DON'T:** Use generic error types; always provide context

### Async/Await
✅ **DO:** Use `async`/`await` consistently with `?` operator
```rust
pub async fn scan_dependencies(&self, sbom_id: Uuid) -> Result<Vec<Vulnerability>> {
    let deps = self.db.get_dependencies(sbom_id).await?;
    let vulns = self.scanner.scan(&deps).await?;
    Ok(vulns)
}
```
❌ **DON'T:** Use `.unwrap()` in production code; always handle errors properly

## API Endpoints

### REST API

**Base URL:** `http://localhost:7106`

**Key Endpoints:**
- `GET /healthz` - Health check
- `GET /readyz` - Readiness check (includes database connectivity)
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/sboms` - Generate SBOM for project
- `GET /api/v1/sboms/{id}` - Get SBOM by ID
- `POST /api/v1/sboms/{id}/scan` - Scan SBOM for vulnerabilities
- `GET /api/v1/vulnerabilities/{id}` - Get vulnerability details
- `GET /api/v1/dependencies/{id}/vulnerabilities` - Get vulnerabilities for dependency
- `POST /api/v1/policies/validate` - Validate dependencies against policies
- `GET /api/v1/projects/{id}/sbom` - Get latest SBOM for project

### gRPC API

**Port:** 50066

**Services:**
- `DepGovernanceService` - Dependency management and scanning
  - `GenerateSBOM` - Generate SBOM for project
  - `ScanVulnerabilities` - Scan dependencies for CVEs
  - `ValidatePolicy` - Check policy compliance
  - `GetDependencyGraph` - Retrieve dependency tree

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`sboms`** - Software Bill of Materials records
- Columns: `id`, `project_id`, `format` (cyclonedx|spdx), `version`, `content` (JSONB), `created_at`
- Indexes: `idx_project_id`, `idx_created_at`
- Purpose: Store generated SBOMs for auditing and compliance

**`dependencies`** - Dependency inventory
- Columns: `id`, `sbom_id`, `name`, `version`, `license`, `package_manager`, `is_direct`, `created_at`
- Indexes: `idx_sbom_id`, `idx_name_version`, `idx_license`
- Purpose: Normalized dependency data for querying and analysis

**`vulnerabilities`** - Known vulnerabilities
- Columns: `id`, `cve_id`, `severity`, `cvss_score`, `affected_packages`, `fixed_version`, `published_at`
- Indexes: `idx_cve_id`, `idx_severity`, `idx_affected_packages`
- Purpose: CVE database for vulnerability matching

**`dependency_vulnerabilities`** - Junction table
- Columns: `dependency_id`, `vulnerability_id`, `detected_at`
- Indexes: `idx_dependency_id`, `idx_vulnerability_id`
- Purpose: Track which dependencies have which vulnerabilities

**Migrations:**
- Location: `migrations/`
- Tool: SQLx (`sqlx migrate`)
- Commands: `sqlx migrate add <name>`, `sqlx migrate run`, `sqlx migrate revert`

## Event Handling

**Published Events:**
- `sbom.generated` - When SBOM created for project
  - Payload: `{sbom_id, project_id, format, dependency_count}`
- `vulnerability.detected` - When new vulnerability found in dependencies
  - Payload: `{cve_id, severity, affected_dependencies[], project_ids[]}`
- `policy.violation` - When dependency violates policy
  - Payload: `{policy_id, dependency, violation_type, project_id}`
- `dependency.updated` - When dependency version changed
  - Payload: `{dependency_id, old_version, new_version, has_breaking_changes}`

**Subscribed Events:**
- `project.created` - Initialize dependency tracking for new project
- `deployment.started` - Trigger SBOM generation and vulnerability scan

## Testing Strategy

### Unit Tests
- Location: Inline with `#[cfg(test)]` modules in source files
- Coverage: Target >80%
- Mock: Database with test database, external services with `wiremock`
- Example: `src/services/sbom_generator.rs` has unit tests for SBOM parsing

### Integration Tests
- Location: `tests/` directory
- Setup: Use Testcontainers for PostgreSQL
- Pattern: Test full request flow from HTTP handler -> service -> database

### Running Tests
```bash
# All tests
cargo test --all

# Specific test
cargo test test_sbom_generation

# With output
cargo test -- --nocapture

# Integration tests only
cargo test --test integration_test
```

## Configuration

### Environment Variables
```bash
# Service-specific config
DATABASE_URL=postgresql://user:pass@localhost:5432/depgovernance
SBOM_FORMAT=cyclonedx  # cyclonedx | spdx
SCAN_ENABLED=true
VULNERABILITY_SOURCES=osv,nvd,github
POLICY_MODE=enforce  # enforce | audit
LICENSE_ALLOWLIST=MIT,Apache-2.0,BSD-3-Clause
LICENSE_BLOCKLIST=GPL-3.0,AGPL-3.0

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=dep-governance-service
SERVICE_PORT=7106
GRPC_PORT=50066
ENV=dev
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=dep-governance-service
LOG_LEVEL=info
RUST_LOG=dep_governance_service=debug,sqlx=warn
```

### Secrets
- Stored in: Vault at `secret/dep-governance-service/`
- Accessed via: Vault SDK or secrets-env-service
- Keys: API tokens for vulnerability databases (NVD, GitHub)

## Quick Find Commands

### Find Code
```bash
# Find SBOM generation logic
rg -n "generate.*sbom|create_sbom" services/dep-governance-service/src/

# Find vulnerability scanning
rg -n "scan.*vulnerabilit" services/dep-governance-service/src/

# Find database queries
rg -n "sqlx::query!" services/dep-governance-service/src/db/

# Find event publishers
rg -n "publish.*sbom\.|vulnerability\." services/dep-governance-service/src/events/

# Find policy enforcement
rg -n "PolicyEngine|validate.*policy" services/dep-governance-service/src/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "dep-governance-service" --glob "docker-compose*.yml" --glob "*.yaml"

# Find axum route definitions
rg -n "Router::new|\.route\(" services/dep-governance-service/src/
```

## Common Gotchas

- **SQLx Compile-Time Verification:** `sqlx::query!` requires database connection at compile time; set `DATABASE_URL` or use offline mode with `cargo sqlx prepare`
- **Async Runtime:** All I/O must use `async`/`await`; blocking operations will block entire Tokio thread pool
- **Lifetime Annotations:** Rust borrow checker may require explicit lifetimes; start with compiler suggestions before adding manually
- **Large SBOMs:** Projects with many dependencies generate large SBOM JSON; consider pagination or compression for API responses
- **CVE Database Updates:** Vulnerability data becomes stale; implement periodic sync from upstream sources (NVD, OSV)
- **Ownership vs Borrowing:** Prefer references (`&T`) over cloning (`T.clone()`) for large structs to avoid performance overhead

## Related Services

- **projects-service:** Provides project metadata; Dep Governance queries project configuration for policy settings
- **notification-service:** Sends alerts when critical vulnerabilities detected in dependencies
- **db-guardian-service:** Collaborates on migration validation by checking dependency compatibility
- **observability-service:** Aggregates vulnerability metrics and SBOM generation trends

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-dep-governance-service.yml`
- Rust API Docs: Run `cargo doc --open` to view generated documentation
- Migration Guide: `PHASE1_COMPLETE.md`, `PHASE2_PROGRESS.md`
