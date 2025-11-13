# Secrets Env Service MVP Hardening — Technical Specification

Document Name: secrets-env-service-mvp-hardening-spec
Date: 2025-11-12
Version: v1
Status: Active

## Executive Summary
Productionize the secrets-env-service for MVP by adding OTLP tracing export, application metrics, input validation hardening, JWT-claims–driven RBAC, gRPC health/reflection, and startup readiness, without regressing existing tests or APIs.

## Architecture Overview
- Tracing: OTel TracerProvider with OTLP gRPC exporter (configurable endpoint/TLS), graceful shutdown hook.
- Metrics: Prometheus client_golang counters/histograms for HTTP/gRPC and domain operations, exposed at /metrics.
- Validation/Security: Middleware for Content-Type and body size limits; env allowlist; RBAC roles from JWT claims with header override for dev/tests.
- Operability: gRPC health service and reflection (config toggles); readiness checks for DB/Redis/NATS/Vault at startup with fail-fast or degraded modes per backend.
- Config: New env flags for OTEL exporter, metrics, limits, reflection/health toggles.

## Implementation Phases
1) Traces
   - Add OTLP exporter (gRPC), TLS/insecure toggles, headers; ensure provider shutdown.
2) Metrics
   - HTTP counters and latency histograms by method/route/status; gRPC counters and latencies by method/code; domain operation counters.
3) Validation & Security
   - Env allowlist; enforce JSON Content-Type; request body size cap; RBAC from JWT claims with header override.
4) Operability
   - gRPC reflection + health behind flags; startup readiness for DB/Redis/NATS/Vault.
5) Config & Docs
   - Add config fields and defaults; document new env vars and example commands.

## Testing & Verification
- Unit: metrics via prometheus/testutil; validation (env, content-type, size); RBAC claims handling.
- Integration: assert select HTTP/gRPC metrics; end-to-end spans to collector when OTELEndpoint configured.
- DoD: all existing tests green; traces visible in collector; metrics exposed and labeled; toggles work.

## Security Considerations
- Do not log secrets/PII; validate content types and sizes; prefer roles from signed JWT claims; secure OTLP with TLS by default; mTLS unaffected.
