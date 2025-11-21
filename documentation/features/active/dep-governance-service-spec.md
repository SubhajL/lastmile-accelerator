# dep-governance-service — Feature Spec

- Service path: `services/dep-governance-service`
- Technology: Rust 2021, Axum web framework
- Ports: REST 7106, gRPC 50066 (reserved)
- Links: `services/dep-governance-service/CLAUDE.md`, `services/dep-governance-service/CONTEXT.md`, `service_catalog.yaml`
- Status: **Infrastructure Complete, Core Features Not Implemented**

## Summary

The Dependency Governance Service provides comprehensive Software Bill of Materials (SBOM) generation, dependency tracking, and vulnerability management for all projects in the Last-Mile Accelerator platform. It automates security compliance by parsing package manifests, generating standards-compliant SBOMs (CycloneDX, SPDX), scanning dependencies against CVE databases, and enforcing organizational dependency policies.

**Current Reality:** The service has production-ready infrastructure (75% complete) but **lacks all core business logic** (0% complete). It can store SBOM metadata and CVE records but cannot generate SBOMs, scan for vulnerabilities, or enforce policies.

## Goals

### Implemented ✅
- [x] Production-ready HTTP server with health checks and metrics
- [x] JWT authentication with RS256/JWKS validation
- [x] PostgreSQL connection pooling with migrations
- [x] Database schema for dependencies, SBOMs, CVEs, and vulnerability linkage
- [x] REST API CRUD endpoints for metadata management
- [x] OpenTelemetry tracing and structured logging
- [x] NATS event bus integration (publisher ready)
- [x] Error handling with typed errors
- [x] Unit and integration test infrastructure

### Not Implemented ❌
- [ ] **SBOM Generation** from package manifests (package.json, Cargo.toml, go.mod, pom.xml)
- [ ] **CycloneDX/SPDX** format output
- [ ] **Vulnerability Scanning** via OSV, NVD, GitHub Advisory APIs
- [ ] **CVE Matching** and enrichment logic
- [ ] **Policy Enforcement** (license validation, version constraints)
- [ ] **S3/MinIO Integration** for SBOM file storage
- [ ] **Event-Driven Workflows** (subscribe to project events, publish scan results)
- [ ] **Transitive Dependency Resolution** (build full dependency trees)
- [ ] **gRPC API** implementation
- [ ] **Redis Caching** layer
- [ ] **Background Jobs** for periodic vulnerability database updates

## Non-goals

- **Manual SBOM Upload** - Service generates SBOMs; doesn't accept user-uploaded SBOMs
- **Real-time CVE Monitoring** - Scans on-demand or scheduled; not continuous streaming
- **Proprietary Vulnerability Databases** - Only uses public sources (OSV, NVD, GitHub)
- **License Legal Advice** - Flags license issues but doesn't provide legal recommendations
- **Automatic Dependency Updates** - Recommends updates but doesn't modify code
- **Build System Integration** - Works with manifest files; doesn't hook into build tools directly

## Interfaces

### REST API (Implemented - CRUD Only)

**Base URL:** `http://localhost:7106`

**Health & Observability:**
- `GET /healthz` → 200 OK (liveness)
- `GET /readyz` → 200 OK if DB connected (readiness)
- `GET /metrics` → Prometheus format metrics

**SBOM Management (Metadata Only):**
- `POST /v1/snapshots/:snapshot_id/sbom` → Create SBOM metadata record
  - Body: `{format: "cyclonedx" | "spdx_json" | "spdx_xml", storage_key: string, file_hash: string}`
  - Returns: 201 Created with SBOM ID
- `GET /v1/snapshots/:snapshot_id/sbom` → Get latest SBOM for snapshot
  - Returns: 200 OK with SBOM metadata (NOT the actual SBOM file)

**Dependency Management:**
- `GET /v1/snapshots/:snapshot_id/dependencies?direct={true|false}` → List dependencies
  - Query param `direct=true` filters to direct dependencies only
  - Returns: JSON array of dependencies

**Vulnerability Management:**
- `POST /v1/cves` → Upsert CVE record
  - Body: `{cve_id: string, severity: string, description: string?, ...}`
  - Returns: 200 OK with CVE ID
- `POST /v1/dependencies/:dependency_id/vulns/link` → Link CVE to dependency
  - Body: `{cve_id: UUID, status: string, affected_versions: string?, fixed_version: string?}`
  - Returns: 201 Created
- `GET /v1/dependencies/:dependency_id/vulns` → Get vulnerabilities for dependency
  - Returns: JSON array with joined CVE data

**Authentication:**
- All `/v1/*` routes require `Authorization: Bearer <jwt-token>`
- JWT must be RS256 signed, validated via JWKS endpoint
- Required claims: `sub`, `iss`, `aud`, `exp`

### REST API (Not Implemented)

**SBOM Generation:**
- `POST /v1/projects/:project_id/sbom/generate` → Generate SBOM from repository
  - Body: `{source: "git" | "archive", location: string, format: "cyclonedx" | "spdx"}`
  - Should: Clone repo, detect package managers, parse manifests, generate SBOM, upload to S3, store metadata
  - Returns: 202 Accepted with job ID

**Vulnerability Scanning:**
- `POST /v1/sboms/:sbom_id/scan` → Scan SBOM for vulnerabilities
  - Should: Query OSV/NVD for each dependency, enrich with CVE data, store results
  - Returns: 202 Accepted with scan ID
- `GET /v1/scans/:scan_id/status` → Check scan progress
  - Returns: `{status: "pending" | "running" | "completed" | "failed", vulnerabilities_found: number}`

**Policy Validation:**
- `POST /v1/sboms/:sbom_id/validate` → Validate against org policies
  - Body: `{policy_id?: UUID}` (optional specific policy)
  - Should: Check licenses, version constraints, blocklist
  - Returns: `{passed: boolean, violations: [{type, dependency, message}]}`

### gRPC API (Not Implemented)

**Port:** 50066

**Proto File (Planned):**
```protobuf
service DepGovernanceService {
  rpc GenerateSBOM(GenerateSBOMRequest) returns (GenerateSBOMResponse);
  rpc ScanVulnerabilities(ScanRequest) returns (ScanResponse);
  rpc ValidatePolicy(ValidatePolicyRequest) returns (ValidatePolicyResponse);
  rpc GetDependencyGraph(GetGraphRequest) returns (GetGraphResponse);
}
```

### Event Bus (NATS)

**Published Events (Not Implemented):**
- `sbom.generated` → `{sbom_id, project_id, format, dependency_count, generated_at}`
- `vulnerability.detected` → `{cve_id, severity, affected_dependencies[], project_ids[], detected_at}`
- `policy.violation` → `{policy_id, violation_type, dependency, project_id, severity}`
- `dependency.updated` → `{dependency_id, old_version, new_version, has_breaking_changes}`

**Subscribed Events (Not Implemented):**
- `project.created` → Trigger initial SBOM generation
- `deployment.started` → Trigger vulnerability scan before deploy
- `code.pushed` → Regenerate SBOM if manifest changed

## Dependencies

### Internal Services
- **projects-service** → Queries project metadata (language, repository URL)
- **notification-service** → Sends alerts for critical vulnerabilities
- **secrets-env-service** → Retrieves API keys for NVD, GitHub
- **observability-service** → Aggregates vulnerability metrics

### External Services
- **PostgreSQL** → Primary data store (dependencies, SBOMs, CVEs)
- **Redis** → Caching layer (vulnerability data, SBOM metadata) - configured but not used
- **NATS** → Event bus for async workflows - publisher ready, no subscribers
- **S3/MinIO** → Object storage for actual SBOM files - not integrated
- **Vault** → Secrets management for API keys - config present, not used

### External APIs (Not Integrated)
- **OSV (Open Source Vulnerabilities)** → https://api.osv.dev/v1/query
  - Free, public API for vulnerability data
  - Supports npm, cargo, pypi, maven, go, etc.
- **NVD (National Vulnerability Database)** → https://services.nvd.nist.gov/rest/json/cves/2.0
  - NIST-maintained CVE database
  - Requires free API key (rate limited)
- **GitHub Advisory Database** → https://api.github.com/advisories
  - GitHub's security advisories
  - Part of GitHub REST API

## Data Models

### Database Schema (Implemented)

**`dependencies` table:**
```sql
CREATE TABLE dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    ecosystem TEXT NOT NULL, -- npm, cargo, pypi, maven, go, etc.
    is_direct BOOLEAN NOT NULL DEFAULT false,
    parent_id UUID REFERENCES dependencies(id), -- for transitive deps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(snapshot_id, name, version, ecosystem)
);
CREATE INDEX idx_dependencies_snapshot ON dependencies(snapshot_id);
CREATE INDEX idx_dependencies_snapshot_direct ON dependencies(snapshot_id, is_direct);
CREATE INDEX idx_dependencies_name_gin ON dependencies USING gin(name gin_trgm_ops);
CREATE INDEX idx_dependencies_parent ON dependencies(parent_id);
```

**`sboms` table:**
```sql
CREATE TABLE sboms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL,
    format TEXT NOT NULL CHECK (format IN ('cyclonedx', 'spdx_json', 'spdx_xml')),
    storage_key TEXT NOT NULL, -- S3 object key
    file_hash TEXT NOT NULL, -- SHA-256 hash for verification
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_sboms_snapshot ON sboms(snapshot_id);
CREATE INDEX idx_sboms_created ON sboms(created_at DESC);
```

**`cves` table:**
```sql
CREATE TABLE cves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cve_id TEXT UNIQUE NOT NULL, -- e.g., "CVE-2023-12345"
    severity TEXT NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    description TEXT,
    published_at TIMESTAMP WITH TIME ZONE,
    source TEXT, -- "osv", "nvd", "github"
    cvss_score REAL CHECK (cvss_score >= 0.0 AND cvss_score <= 10.0),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_cves_cve_id ON cves(cve_id);
CREATE INDEX idx_cves_severity ON cves(severity);
```

**`dependency_vulnerabilities` table (junction):**
```sql
CREATE TABLE dependency_vulnerabilities (
    dependency_id UUID REFERENCES dependencies(id) ON DELETE CASCADE,
    cve_id UUID REFERENCES cves(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'open', -- open, acknowledged, fixed, ignored, false_positive
    affected_versions TEXT, -- semver range like ">=1.0.0 <1.2.0"
    fixed_version TEXT,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (dependency_id, cve_id)
);
CREATE INDEX idx_depvulns_dependency ON dependency_vulnerabilities(dependency_id);
CREATE INDEX idx_depvulns_cve ON dependency_vulnerabilities(cve_id);
```

### Rust Types (Implemented)

**Common Enums:**
```rust
pub enum Ecosystem {
    Npm, Cargo, Pypi, Maven, Go, Nuget, Composer, Rubygems
}

pub enum Severity {
    Critical, High, Medium, Low, Info
}

pub enum LicenseType {
    Permissive, Copyleft, Proprietary, Unknown
}

pub enum ScanStatus {
    Pending, Running, Completed, Failed
}

pub enum VulnerabilityStatus {
    Open, Acknowledged, Fixed, Ignored, FalsePositive
}
```

**Core Models:**
```rust
pub struct Dependency {
    pub id: Uuid,
    pub snapshot_id: Uuid,
    pub name: String,
    pub version: String,
    pub ecosystem: Ecosystem,
    pub is_direct: bool,
    pub parent_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
}

pub struct Sbom {
    pub id: Uuid,
    pub snapshot_id: Uuid,
    pub format: String, // "cyclonedx", "spdx_json", "spdx_xml"
    pub storage_key: String,
    pub file_hash: String,
    pub created_at: DateTime<Utc>,
}

pub struct Cve {
    pub id: Uuid,
    pub cve_id: String,
    pub severity: String,
    pub description: Option<String>,
    pub published_at: Option<DateTime<Utc>>,
    pub source: Option<String>,
    pub cvss_score: Option<f32>,
    pub updated_at: DateTime<Utc>,
}
```

## Risks

### Technical Risks

**SBOM Generation Complexity:**
- **Risk:** Parsing diverse package manifests (npm, cargo, maven, go) requires ecosystem-specific logic
- **Mitigation:** Start with npm (most common); delegate complex cases to Syft tool
- **Status:** Not mitigated; no implementation started

**Vulnerability Data Staleness:**
- **Risk:** CVE databases update frequently; local cache becomes outdated
- **Mitigation:** Implement periodic background sync (daily); allow manual refresh
- **Status:** Not mitigated; no caching implemented

**API Rate Limits:**
- **Risk:** NVD API has strict rate limits (50 requests/30 seconds without key)
- **Mitigation:** Batch queries; cache results; use OSV as primary source
- **Status:** Not mitigated; no API integration

**Transitive Dependency Explosion:**
- **Risk:** Large projects have 1000+ transitive dependencies; scanning is slow
- **Mitigation:** Async job processing; pagination; focus on direct deps first
- **Status:** Not mitigated; no transitive resolution

### Operational Risks

**Storage Costs:**
- **Risk:** SBOM files are large (100KB-10MB); storing all versions is expensive
- **Mitigation:** Retention policy (90 days); compress old SBOMs; deduplicate identical files
- **Status:** Not addressed; S3 integration not implemented

**False Positives:**
- **Risk:** CVE matching may flag non-applicable vulnerabilities (wrong platform, dev-only deps)
- **Mitigation:** Allow users to mark false positives; filter dev dependencies by default
- **Status:** Status field exists in schema; no UI or logic implemented

**Policy Conflicts:**
- **Risk:** Multiple policies may conflict (org-level vs project-level)
- **Mitigation:** Clear precedence rules (project > team > org); policy versioning
- **Status:** No policy implementation started

### Security Risks

**SBOM Information Disclosure:**
- **Risk:** SBOM reveals dependency versions attackers can target
- **Mitigation:** Require authentication for SBOM access; audit logs
- **Status:** Partially mitigated; JWT auth enforced, but no audit logs

**Malicious Package Manifests:**
- **Risk:** Crafted manifest files could exploit parser vulnerabilities
- **Mitigation:** Sandboxed parsing; input validation; size limits
- **Status:** Not mitigated; no parsers implemented

**CVE Database Poisoning:**
- **Risk:** Attacker could submit fake CVEs to public databases
- **Mitigation:** Cross-reference multiple sources (OSV + NVD); require verified sources
- **Status:** Not mitigated; no integration with CVE sources

## Milestones

### M1: Infrastructure Setup ✅ **COMPLETE** (Nov 2025)
- ✅ Axum web server with Tokio async runtime
- ✅ PostgreSQL connection pooling (sqlx)
- ✅ JWT authentication with JWKS provider
- ✅ OpenTelemetry tracing and Prometheus metrics
- ✅ NATS event publisher with retry logic
- ✅ Health and readiness checks
- ✅ Graceful shutdown handling
- **Deliverable:** Service starts, responds to health checks, validates JWT tokens

### M2: Data Layer ⏳ **IN PROGRESS** (60% complete)
- ✅ Database migrations (dependencies, sboms, cves, junctions)
- ✅ Common types (Ecosystem, Severity, LicenseType enums)
- ✅ Core data models (Dependency, Sbom, Cve)
- ✅ Repository layer (CRUD operations)
- ✅ REST API handlers (metadata management)
- ✅ Unit tests (31 passing)
- ❌ SBOM generation logic
- ❌ Vulnerability scanning logic
- **Deliverable:** Can store and retrieve dependency metadata; cannot generate or scan

### M3: SBOM Generation ❌ **NOT STARTED** (Est. 12-16 hours)
- ❌ npm package.json parser
- ❌ Cargo.toml parser (Rust)
- ❌ CycloneDX 1.5 JSON generator
- ❌ SPDX 2.3 JSON generator
- ❌ Integration tests with real manifest files
- ❌ S3/MinIO client for file upload
- **Deliverable:** Generate CycloneDX SBOM from package.json; upload to S3

### M4: Vulnerability Scanning ❌ **NOT STARTED** (Est. 12-16 hours)
- ❌ OSV API client (https://api.osv.dev/v1/query)
- ❌ NVD API client (https://services.nvd.nist.gov/rest/json/cves/2.0)
- ❌ CVE matching logic (dependency name/version → CVE records)
- ❌ Severity calculation (CVSS score → Critical/High/Medium/Low)
- ❌ Background job for periodic vulnerability database updates
- ❌ Event publishing (vulnerability.detected)
- **Deliverable:** Scan SBOM dependencies; detect known CVEs; store results

### M5: Event-Driven Workflows ❌ **NOT STARTED** (Est. 6-8 hours)
- ❌ Subscribe to `project.created` → generate initial SBOM
- ❌ Subscribe to `deployment.started` → run vulnerability scan
- ❌ Publish `sbom.generated` after successful generation
- ❌ Publish `vulnerability.detected` after scan finds issues
- ❌ Async job processing (tokio tasks or background worker)
- **Deliverable:** Automated SBOM generation and scanning triggered by events

### M6: Policy Enforcement ❌ **NOT STARTED** (Est. 8-10 hours)
- ❌ YAML policy file parsing
- ❌ License validation (allowlist/blocklist)
- ❌ Version constraint checking (semver ranges)
- ❌ Security policy rules (min CVSS score threshold)
- ❌ Policy violation reporting
- ❌ Event publishing (policy.violation)
- **Deliverable:** Block deployments with GPL dependencies or critical CVEs

### M7: Additional Languages ❌ **NOT STARTED** (Est. 12-18 hours)
- ❌ go.mod parser (Go)
- ❌ pom.xml parser (Maven)
- ❌ requirements.txt / pyproject.toml parser (Python)
- ❌ composer.json parser (PHP)
- **Deliverable:** Multi-language SBOM generation support

### M8: Production Hardening ❌ **NOT STARTED** (Est. 10-15 hours)
- ❌ gRPC server implementation (port 50066)
- ❌ Redis caching layer (vulnerability data, SBOM metadata)
- ❌ Rate limiting (prevent API abuse)
- ❌ Custom Prometheus metrics (sbom_generation_duration, vulnerabilities_detected_total)
- ❌ Load testing (JMeter, k6)
- ❌ Security audit (penetration testing)
- **Deliverable:** Production-ready service with performance benchmarks

## Current Implementation Gap

### What Works Now (As of Nov 2025)
```bash
# Start service
SERVICE_PORT=7106 DATABASE_URL=postgresql://... ./target/release/dep_governance_service

# Health check (works)
curl http://localhost:7106/healthz
# → {"status": "ok"}

# Create SBOM metadata (works, but doesn't actually generate SBOM)
curl -X POST http://localhost:7106/v1/snapshots/123e4567-e89b-12d3-a456-426614174000/sbom \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"format": "cyclonedx", "storage_key": "sboms/project-abc/sbom-123.json", "file_hash": "sha256..."}'
# → 201 Created

# Get SBOM metadata (works, but returns metadata only, not actual SBOM)
curl http://localhost:7106/v1/snapshots/123e4567-e89b-12d3-a456-426614174000/sbom \
  -H "Authorization: Bearer <jwt>"
# → {"id": "...", "format": "cyclonedx", "storage_key": "...", "created_at": "..."}
```

### What Doesn't Work (Critical Gaps)
```bash
# Generate SBOM from repository (NOT IMPLEMENTED)
curl -X POST http://localhost:7106/v1/projects/abc/sbom/generate \
  -H "Authorization: Bearer <jwt>" \
  -d '{"source": "git", "location": "https://github.com/user/repo", "format": "cyclonedx"}'
# → 404 Not Found (endpoint doesn't exist)

# Scan for vulnerabilities (NOT IMPLEMENTED)
curl -X POST http://localhost:7106/v1/sboms/xyz/scan \
  -H "Authorization: Bearer <jwt>"
# → 404 Not Found (endpoint doesn't exist)

# Validate against policy (NOT IMPLEMENTED)
curl -X POST http://localhost:7106/v1/sboms/xyz/validate \
  -H "Authorization: Bearer <jwt>"
# → 404 Not Found (endpoint doesn't exist)
```

## Next Steps to Reach MVP

### Phase 1: Quick Wins (1-2 hours) - **Recommended First**
1. Run `cargo fmt` to fix formatting issues
2. Fix clippy warning (move test module to end of file in `src/handlers/api/sbom.rs`)
3. Commit current work: `git add -A && git commit -m "docs: update progress and spec for dep-governance-service"`
4. Push to remote: `git push origin auth-validation`
5. Update README.md with accurate capabilities

### Phase 2: CI/CD Setup (2-3 hours) - **High Priority**
1. Create `.github/workflows/ci-dep-governance-service.yml`
2. Configure PR gates: fmt check, clippy, tests, build
3. Configure push-to-main gates (same as PR)
4. Test workflow on feature branch

### Phase 3: SBOM Generation (12-16 hours) - **MVP Critical**
1. Implement npm `package.json` parser
2. Generate CycloneDX 1.5 JSON format
3. Integrate S3/MinIO client (aws-sdk-s3)
4. Add `POST /v1/projects/:id/sbom/generate` endpoint
5. Write integration tests with real package.json files

### Phase 4: Vulnerability Scanning (12-16 hours) - **MVP Critical**
1. Implement OSV API client
2. Add CVE matching logic
3. Add `POST /v1/sboms/:id/scan` endpoint
4. Store vulnerability records in database
5. Publish `vulnerability.detected` events

### Phase 5: Event Workflows (6-8 hours) - **MVP Important**
1. Subscribe to `project.created` event
2. Subscribe to `deployment.started` event
3. Trigger SBOM generation + scan on events
4. Add async job processing

### Phase 6: Policy Engine (8-10 hours) - **MVP Nice-to-Have**
1. YAML policy file parsing
2. License allowlist/blocklist validation
3. CVSS threshold enforcement
4. Policy violation reporting

---

**Total Estimated Time to Functional MVP:** 40-60 hours
**Current Progress:** 75% infrastructure, 0% business logic
**Blocker Status:** None; all dependencies and APIs available
**Recommendation:** Prioritize Phases 1-4; defer Phase 6 to post-MVP
