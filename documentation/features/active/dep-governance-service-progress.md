# dep-governance-service — Progress

- Service path: `services/dep-governance-service`
- Branch: `auth-validation`
- Last Updated: 2025-11-21

## Status: Infrastructure Complete ✅ | Core Features Not Started ❌

### Overall Progress: 75% Infrastructure, 0% Business Logic

## Checklist
- [x] Design - Infrastructure phase complete
- [x] Implementation - Phase 1 (Infrastructure) complete; Phase 2 (Data models & DB) ~60% complete
- [x] Unit tests - Infrastructure tests passing (31 tests)
- [x] Integration tests - HTTP handler tests implemented (4 test files)
- [x] Observability - OpenTelemetry, structured logging, health checks complete
- [x] Helm deploy - Helm charts in place
- [ ] **SBOM Generation** - NOT IMPLEMENTED (core feature)
- [ ] **Vulnerability Scanning** - NOT IMPLEMENTED (core feature)
- [ ] **Policy Enforcement** - NOT IMPLEMENTED (core feature)
- [ ] **S3/MinIO Integration** - NOT IMPLEMENTED
- [ ] **gRPC API** - NOT IMPLEMENTED

## Current State Summary

### ✅ What's Complete (Phase 1 + 2 Foundation)

**Infrastructure (100%)**
- ✅ Configuration management with AppConfig
- ✅ PostgreSQL connection pooling (5-20 connections)
- ✅ Redis client configuration
- ✅ NATS event bus with retry logic
- ✅ HTTP server (Axum) with middleware
- ✅ Health endpoints: `/healthz`, `/readyz`, `/metrics`
- ✅ JWT authentication middleware (RS256, JWKS provider)
- ✅ OpenTelemetry tracing and structured logging
- ✅ Graceful shutdown handling
- ✅ Error handling with typed errors

**Data Layer (60%)**
- ✅ Database migrations (3 migrations):
  - `dependencies` table (UUID, tree structure, indexes)
  - `sboms` table (metadata storage)
  - `cves` and `dependency_vulnerabilities` tables
- ✅ Common types module (Ecosystem, Severity, LicenseType, etc.)
- ✅ Data models: Dependency, Sbom, Cve, DependencyVulnerability
- ✅ Database repositories:
  - SBOM CRUD operations
  - Dependency listing and filtering
  - CVE upsert and linking
  - Vulnerability queries with joins

**REST API (CRUD Only)**
- ✅ `POST /v1/snapshots/:snapshot_id/sbom` - Create SBOM metadata record
- ✅ `GET /v1/snapshots/:snapshot_id/sbom` - Get latest SBOM
- ✅ `GET /v1/snapshots/:snapshot_id/dependencies?direct=true` - List dependencies
- ✅ `GET /v1/dependencies/:dependency_id/vulns` - Get vulnerabilities
- ✅ `POST /v1/cves` - Upsert CVE
- ✅ `POST /v1/dependencies/:dependency_id/vulns/link` - Link CVE to dependency

**Testing (31 tests passing)**
- ✅ Unit tests for common types (10 tests)
- ✅ Infrastructure tests (17 tests)
- ✅ Integration tests for HTTP handlers (4 test files)

**Build & Deployment**
- ✅ Makefile with build, test, run, image targets
- ✅ Dockerfile (multi-stage with distroless)
- ✅ Helm chart structure
- ✅ Release binary builds successfully (9.9MB)

### ❌ What's Missing for MVP

**Core Business Logic (0%)**
1. ❌ **SBOM Generation Service** - Parse package.json, Cargo.toml, go.mod, pom.xml and generate CycloneDX/SPDX format SBOMs
2. ❌ **Vulnerability Scanner Service** - Query OSV, NVD, GitHub Advisory for CVEs; match dependencies to vulnerabilities
3. ❌ **Policy Engine Service** - Validate licenses, enforce version constraints, check security policies
4. ❌ **Dependency Tree Builder** - Recursive resolution of transitive dependencies with cycle detection
5. ❌ **Event Publishing** - Publish `sbom.generated`, `vulnerability.detected`, `policy.violation` events
6. ❌ **Event Subscription** - Subscribe to `project.created`, `deployment.started` events

**Storage & External Integrations (0%)**
7. ❌ **S3/MinIO Integration** - Upload actual SBOM files (currently only metadata in DB)
8. ❌ **OSV API Client** - Query https://api.osv.dev/v1/query for vulnerabilities
9. ❌ **NVD API Client** - Query https://services.nvd.nist.gov/rest/json/cves/2.0 for CVE data
10. ❌ **GitHub Advisory Client** - Query GitHub Security Advisory Database

**Advanced Features (0%)**
11. ❌ **gRPC Server** - Port 50066 reserved but not implemented
12. ❌ **Redis Caching** - Client configured but never used
13. ❌ **Background Jobs** - No async job processing for periodic vulnerability updates
14. ❌ **Transitive Dependency Resolution** - Cannot build full dependency trees

### Code Quality Status

**Strengths:**
- ✅ Compiles successfully (release build works)
- ✅ Clean separation: handlers → services → repositories → database
- ✅ Strong type safety with Rust ownership model
- ✅ Comprehensive enum types with Display/FromStr traits
- ✅ SQLx compile-time query verification
- ✅ Code formatting PASSES: `cargo fmt --check` shows no issues (FIXED since last update)
- ✅ Clippy PASSES: No warnings in production code (FIXED since last update)

**Issues to Fix:**
- ⚠️ **6 test failures**: Config validation tests failing (43 tests passing, 6 failing, 1 ignored)
  - `test_config_defaults_applied_correctly`
  - `test_config_fails_without_required_vars`
  - `test_config_loads_from_env`
  - `test_config_validates_database_url_format`
  - `test_config_validates_nats_url_format`
  - `test_config_validates_port_range`
- ⚠️ **Future incompatibility warnings**: `redis v0.24.0`, `sqlx-postgres v0.7.4` will be rejected by future Rust versions

### Local vs Remote Sync

**Branch Status:**
- Current branch: `auth-validation`
- ✅ **IN SYNC** with `origin/auth-validation` (as of 2025-11-21)
- Uncommitted changes: `.claude/settings.json` (modified), `.claude/hooks/` (untracked)
- Remote: `git@github.com:SubhajL/lastmile-accelerator.git`

**Recent Work (last 11 commits):**
- Comprehensive CLAUDE.md documentation system
- JWT authentication with JWKS provider
- Auth enforcement on /v1 routes with RS256 verification
- Test secret scrubbing for CI compliance

### CI/CD Gates Status

**Current State:**
- ❌ **No service-specific CI workflow found**
- ❌ **No PR gates configured for this service**
- ❌ **No push-to-main gates configured**

**Expected Gates (from CLAUDE.md):**
```bash
# Quality gates that SHOULD exist:
cargo fmt --check        # FAILS (15+ formatting issues)
cargo clippy -- -D warnings  # WOULD FAIL (1 warning)
cargo test --all --locked    # PASSES (but needs TEST_DATABASE_URL)
cargo build --release       # PASSES
```

**Recommended CI Configuration:**
```yaml
# .github/workflows/ci-dep-governance-service.yml
name: dep-governance-service CI
on:
  pull_request:
    paths:
      - 'services/dep-governance-service/**'
  push:
    branches: [main, db-guardian-service]
    paths:
      - 'services/dep-governance-service/**'

jobs:
  test:
    runs-on: ubuntu-latest
    env:
      DATABASE_URL: postgresql://test:test@localhost:5432/test
      TEST_DATABASE_URL: postgresql://test:test@localhost:5432/test
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_USER: test
          POSTGRES_DB: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - name: Format check
        run: cargo fmt --check
        working-directory: services/dep-governance-service
      - name: Clippy
        run: cargo clippy --all-targets -- -D warnings
        working-directory: services/dep-governance-service
      - name: Run migrations
        run: cargo install sqlx-cli && sqlx migrate run
        working-directory: services/dep-governance-service
      - name: Test
        run: cargo test --all --locked
        working-directory: services/dep-governance-service
      - name: Build
        run: cargo build --release
        working-directory: services/dep-governance-service
```

## Milestones

### M1: Infrastructure Setup ✅ COMPLETE
- ✅ Axum server with health endpoints
- ✅ Database connection pooling
- ✅ JWT authentication middleware
- ✅ OpenTelemetry tracing
- ✅ NATS event bus integration
- ✅ Error handling framework

### M2: Data Models & Database ⏳ IN PROGRESS (60%)
- ✅ Common types module
- ✅ Database migrations (3 tables)
- ✅ Data models (Dependency, Sbom, Cve)
- ✅ Repository layer (CRUD operations)
- ❌ SBOM generation logic
- ❌ Vulnerability scanning logic
- ❌ Policy validation logic

### M3: Core Business Logic ❌ NOT STARTED (0%)
- ❌ Package manifest parsers (package.json, Cargo.toml, go.mod, pom.xml)
- ❌ CycloneDX/SPDX format generation
- ❌ OSV/NVD/GitHub Advisory API clients
- ❌ CVE matching and enrichment
- ❌ License policy validation
- ❌ Version constraint checking

### M4: Storage & Events ❌ NOT STARTED (0%)
- ❌ S3/MinIO integration for SBOM files
- ❌ Event publishing (sbom.generated, vulnerability.detected)
- ❌ Event subscription (project.created, deployment.started)
- ❌ Background job system for periodic scans

### M5: Testing & Documentation ⏳ PARTIAL (30%)
- ✅ Unit tests for types and infrastructure
- ✅ Integration tests for HTTP handlers
- ❌ Integration tests for SBOM generation
- ❌ Integration tests for vulnerability scanning
- ❌ OpenAPI/Swagger specification
- ❌ Example policy files and usage documentation

### M6: Production Readiness ❌ NOT STARTED (0%)
- ❌ gRPC server implementation
- ❌ Redis caching layer
- ❌ Rate limiting
- ❌ Prometheus metrics (custom business metrics)
- ❌ Load testing and performance optimization
- ❌ Security audit

## Estimated Work Remaining

### To Reach Functional MVP (40-60 hours)

**Critical Path (Essential for MVP):**
1. **SBOM Generation** (12-16 hours)
   - npm/package.json parser
   - Cargo.toml parser (Rust)
   - CycloneDX JSON output
   - Integration tests with real manifest files

2. **Vulnerability Scanner** (12-16 hours)
   - OSV API client
   - NVD API client (optional, has rate limits)
   - CVE matching logic
   - Background sync for vulnerability database

3. **S3 Storage Integration** (6-8 hours)
   - aws-sdk-s3 or rusoto integration
   - Upload generated SBOMs
   - Presigned URL generation for downloads

4. **Event-Driven Workflows** (6-8 hours)
   - Publish events after SBOM creation
   - Publish events after vulnerability detection
   - Subscribe to project lifecycle events

**Nice-to-Have (Not MVP Blocking):**
5. **Policy Engine** (8-10 hours) - Can be manual validation initially
6. **gRPC API** (6-8 hours) - REST API sufficient for MVP
7. **Redis Caching** (4-6 hours) - Performance optimization
8. **Additional Language Support** (4-6 hours each for Go, Maven, Python)

### Quick Wins (1-2 hours)
- ✅ Fix formatting issues: `cargo fmt`
- ✅ Fix clippy warning (move test module to end of file)
- ✅ Commit and push current work
- ✅ Create CI workflow file
- ✅ Update README.md with current capabilities

## Notes

### Architecture Decisions Made
1. **Database-first approach**: Store SBOM metadata in PostgreSQL, actual files in S3
2. **Async-first**: All I/O uses Tokio async runtime
3. **Strong typing**: Rust enums for all categorical data (Ecosystem, Severity, etc.)
4. **Compile-time safety**: SQLx compile-time query verification
5. **Event-driven**: NATS for loose coupling with other services

### Known Limitations
- SBOM files not actually generated (only metadata stored)
- No vulnerability scanning (only manual CVE entry)
- No policy enforcement (only data storage)
- No transitive dependency resolution
- No automated workflows (no event subscribers)

### Technical Debt
- Update `redis` crate to v0.25+ (current v0.24 has future incompatibility)
- Update `sqlx` to v0.8+ (current v0.7.4 has future incompatibility)
- Add OpenAPI/Swagger spec generation
- Add property-based tests for parsers
- Add load testing for API endpoints

### Blockers
- None currently; all dependencies available
- OSV API is public and free
- NVD API requires free API key (easy to obtain)
- S3-compatible storage available (MinIO in dev stack)

---

## Critical MVP Gap Analysis (2025-11-21)

### Codex Deep Dive Findings

**Infrastructure Status: ✅ 100% Complete**
- HTTP server with Axum + Tokio runtime
- JWT authentication with JWKS provider and caching
- PostgreSQL connection pooling + migrations (sqlx)
- NATS event bus connection (wired but unused)
- OpenTelemetry tracing + Prometheus metrics
- Health/readiness/metrics endpoints
- Error handling with typed errors

**Business Logic Status: ❌ 0% Complete**
- **SBOM Generation:** Only metadata CRUD exists. No manifest parsers (package.json, Cargo.toml, etc.), no CycloneDX/SPDX generators, no S3/MinIO client integration.
- **Vulnerability Scanning:** Manual CVE upsert only. No OSV/NVD/GitHub API clients, no automated CVE matching, no background sync.
- **Policy Enforcement:** Enum types exist (`LicenseType`, `PolicyScope`) but completely unused. No license validation, no version constraint checks.
- **Dependency Resolution:** Read-only listing via API. No manifest ingestion, no transitive dependency graphs.
- **Event-Driven Workflows:** NATS publisher exists but nothing publishes or subscribes. No handlers for `project.created` or `deployment.started`.

**Service Stubs:**
- `src/services/mod.rs`: Placeholder file with comment "Services will be implemented in Phase 2+"
- Event handlers: Wired but unused (`main.rs` passes `_nats`, no consumers)

**Test Status:**
- 43 tests passing (infrastructure, handlers, DB)
- 6 tests failing (config validation edge cases)
- 1 test ignored
- No tests for core business logic (because it doesn't exist)

### MVP Reality Check

**Current Reality:** This is a **beautifully architected shell** with production-ready infrastructure but **zero business value**. It's a REST API that stores SBOM metadata but cannot generate SBOMs, detect vulnerabilities, or enforce policies.

**MVP Requirements NOT Met:**
1. ❌ Cannot generate SBOMs from repositories
2. ❌ Cannot scan dependencies for vulnerabilities
3. ❌ Cannot enforce dependency policies
4. ❌ Cannot react to project lifecycle events
5. ❌ Cannot store actual SBOM files (only metadata)

**What Works:**
- Service starts and responds to health checks
- JWT tokens are validated correctly
- SBOM metadata records can be manually created/retrieved
- CVE records can be manually entered
- Dependencies can be manually inserted and queried

**Estimated Work to Functional MVP:** 40-60 hours of pure implementation
- SBOM Generation: 12-16 hours
- Vulnerability Scanning: 12-16 hours
- S3 Integration: 6-8 hours
- Event Workflows: 6-8 hours
- Policy Engine: 8-10 hours (optional for MVP)

---

**Next Actions:**
1. ✅ Fix code formatting: `cargo fmt` (COMPLETED - passes)
2. ✅ Fix clippy warning in sbom.rs (COMPLETED - passes)
3. ✅ Commit current work: `git add -A && git commit` (COMPLETED - in sync with remote)
4. ✅ Push to remote: `git push origin auth-validation` (COMPLETED - in sync)
5. ❌ Fix 6 failing config tests
6. ❌ Create CI workflow file (`.github/workflows/ci-dep-governance-service.yml`)
7. ❌ Implement SBOM generation (start with npm/package.json parser)
8. ❌ Implement OSV vulnerability scanner API client
9. ❌ Add S3/MinIO storage integration with aws-sdk-s3
10. ❌ Wire up event subscriptions and publishers
