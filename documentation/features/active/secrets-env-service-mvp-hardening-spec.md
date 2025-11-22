# Secrets Env Service MVP Hardening — Technical Specification

Document Name: secrets-env-service-mvp-hardening-spec
Date: 2025-11-12
Version: v2 (Updated 2025-11-21 post-implementation)
Status: Completed

## Executive Summary
Productionize the secrets-env-service for MVP by adding OTLP tracing export, application metrics, input validation hardening, JWT-claims–driven RBAC, gRPC health/reflection, and startup readiness, without regressing existing tests or APIs.

## Architecture Overview
- Tracing: OTel TracerProvider with OTLP gRPC exporter (configurable endpoint/TLS), graceful shutdown hook.
- Metrics: Prometheus client_golang counters/histograms for HTTP/gRPC and domain operations, exposed at /metrics.
- Validation/Security: Middleware for Content-Type and body size limits; env allowlist; RBAC roles from JWT claims (pure JWT, no header override).
- Operability: gRPC health service and reflection (config toggles); readiness checks for DB/Redis/NATS/Vault at startup with fail-fast or degraded modes per backend.
- Config: New env flags for OTEL exporter, metrics, limits, reflection/health toggles.

## Implementation Phases
1) Traces
   - Add OTLP exporter (gRPC), TLS/insecure toggles, headers; ensure provider shutdown.
   - **Implemented**: commits 3700f703, 1c24b5b5, 83aaa7e6
2) Metrics
   - HTTP counters and latency histograms by method/route/status; gRPC counters and latencies by method/code; domain operation counters.
   - **Implemented**: commits 55e96fa4, 31401574, 7b95d0b8
3) Validation & Security
   - Env allowlist; enforce JSON Content-Type; request body size cap; RBAC from JWT claims.
   - **Implementation Deviation**: Removed header override for RBAC (more secure). JWT claims only.
   - **Implemented**: commits c9d226f0, 775ec071, 7387018f, 124e4425
4) Operability
   - gRPC reflection + health behind flags; startup readiness for DB/Redis/NATS/Vault.
   - **Implemented**: commits 1da4292a, 895b1406
5) Config & Docs
   - Add config fields and defaults; document new env vars and example commands.
   - **Enhancement**: Added doc-lint tests to validate README accuracy.
   - **Implemented**: commit 19a34868

## Testing & Verification
- Unit: metrics via prometheus/testutil; validation (env, content-type, size); RBAC claims handling.
- Integration: assert select HTTP/gRPC metrics; end-to-end spans to collector when OTELEndpoint configured.
- DoD: all existing tests green; traces visible in collector; metrics exposed and labeled; toggles work.
- **Result**: All tests passing, 70-90% coverage across critical modules.

## Security Considerations
- Do not log secrets/PII; validate content types and sizes; roles from signed JWT claims only (no dev header override); secure OTLP with TLS by default; mTLS unaffected.

## Implementation Deviations from Original Spec

### 1. RBAC Header Override Removed
**Original Spec**: "RBAC from JWT claims with header override for dev/tests"
**Actual Implementation**: Pure JWT claims only, no header override
**Rationale**: Improved security posture by eliminating potential bypass mechanism
**Impact**: Test environments must use valid JWT tokens (preferred for production parity)

### 2. Domain Metrics Enhancement
**Original Spec**: HTTP/gRPC metrics only
**Actual Implementation**: Added domain-level counters (secrets operations, parity checks, leak scans)
**Rationale**: Better observability into business-level operations
**Impact**: More granular metrics available for monitoring

### 3. Doc Lint Testing
**Original Spec**: Not specified
**Actual Implementation**: Added `internal/doclint` package to validate README env var documentation
**Rationale**: Ensure documentation stays in sync with code
**Impact**: Automated doc quality checks in CI

## Final Configuration Reference

All environment variables implemented as specified in README.md:

### Tracing
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint (http(s)://host:port)
- `OTEL_INSECURE`: Allow insecure OTLP transport (default: false)
- `OTEL_HEADERS`: Comma key=value list sent as headers
- `OTEL_SERVICE_NAME`: Override OTEL service name

### Metrics
- `/metrics` endpoint exposes Prometheus format metrics
- HTTP: `http_requests_total`, `http_request_duration_seconds` (by method, path, status)
- gRPC: `grpc_server_started_total`, `grpc_server_handled_total`, `grpc_server_handling_seconds` (by method, code)
- Domain: `secrets_operations_total`, `parity_checks_total`, `leakscan_operations_total`

### Validation & Security
- `ALLOWED_ENVS`: Env name allowlist (default: "dev,staging,prod,production")
- `HTTP_MAX_BODY_BYTES`: Request body limit (default: 1048576)
- Content-Type enforcement: Requires `application/json` for POST/PUT/PATCH
- JWT verification via `JWT_PUBLIC_KEY` (JWKS URL)
- Roles: `admin` (full access), `auditor` (read-only)

### Operability
- `GRPC_HEALTH_ENABLED`: Enable gRPC health service (default: false)
- `GRPC_REFLECTION_ENABLED`: Enable gRPC reflection (default: false)
- `STARTUP_CRITICAL_TIMEOUT_S`: Timeout for critical checks (default: 3)
- `STARTUP_OPTIONAL_TIMEOUT_S`: Timeout for optional checks (default: 1)
- `/healthz`: Basic process health
- `/readyz`: Readiness (DB/Vault must be healthy)

## Completion Status
✅ **All phases complete** as of 2025-11-21
- Implementation verified via code review and test coverage analysis
- All unit and integration tests passing
- Documentation up to date and validated via doc-lint tests
- Ready for production deployment
