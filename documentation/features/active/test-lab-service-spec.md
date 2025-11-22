# Test Lab Service — Technical Specification

Document Name: test-lab-service Implementation Plan

Date: 2025-11-21

Version: 1.0

Status: Active

## Executive Summary

Test Lab Service provides automated test scaffolding, ephemeral preview environments, and cross-browser execution at scale. It exposes a REST API on port 7202 (Fastify/Node) and plans a gRPC surface on 50072. Core integrations: Postgres (persistence), Redis (cache/rate limit), NATS JetStream (events), S3 (artifacts), Kubernetes Jobs (isolated runs), Selenium Grid/BrowserStack (browser grid). Observability uses OpenTelemetry (traces) and Prometheus (metrics). SLOs: API p95 ≤ 150ms and 99.9% availability. Current implementation is skeletal; see Reality Check for gaps against this plan.

## Architecture Overview

- Application: Node.js + TypeScript (Fastify).
- Middleware: JWT auth (scopes), global error handler, metrics/health endpoints.
- Observability: OTel NodeSDK + auto-instrumentations (HTTP, Fastify, pg, Redis), Prometheus /metrics.
- Data Layer: Postgres with migrations; repos for scaffolds, test_runs, browser_test_runs, preview_environments.
- Execution:
  - K8s runner orchestrates test runs via Jobs; artifacts uploaded to S3.
  - Browser grid runner uses selenium-webdriver against GRID_URL; screenshots/logs to S3.
- Events: Publishers/subscribers over NATS; subjects for run and browser lifecycle; subscribers are feature-flagged.

## Implementation Phases

1) Foundations: Partial — config/logger/error handler present; telemetry initializes globally; /metrics handler currently fails under Bun/vitest with ERR_HTTP_HEADERS_SENT.
2) AuthN/Z: Partial — JWKS verifier + requireAuth/requireScopes/optionalAuth exist; registerAuthPlugin only decorates request.user; legacy `authenticateRequest` removed; tenant access not started.
3) Data: Mostly — migrations + PG repositories for scaffolds/test runs/browser runs/preview environments; pg-mem support used in unit tests; memory backend only supports scaffolds.
4) REST APIs: Partial — routes for scaffolds/test runs/browser runs/previews defined; integration tests currently 500 due to Fastify reply errors and auth mismatch; orchestration triggers absent; test-run/preview routes load only when `REPO_BACKEND=pg`.
5) Runners: Stub — K8s job manifest helper and orchestrator stub with placeholder command; not wired to routes/events; artifact upload helper failing under Bun tests.
6) Browser Grid: Stub — Selenium-based runner stub with screenshot upload; not integrated; S3 import issues in tests.
7) Events: Partial — run/browser subjects defined with publishers; subscribers feature-flagged; wiring not validated end-to-end.
8) gRPC: Not started — no proto/server implementation.

## Reality Check vs Spec (2025-11-21)

- Tests: `bun test` shows 45/143 failures (Fastify ERR_HTTP_HEADERS_SENT on routes/metrics, auth middleware drift, JWKS fixtures vs jose verify, S3 import errors, telemetry shutdown expectation). `bun run typecheck` passes.
- Orchestration: creating test runs does not trigger runner/browser flows or publish events; preview lifecycle absent; no rate limiting or tenant access enforcement.
- Platform/CI: No .github workflows; Makefile `run` points to missing `./bin/server`; Dockerfile unverified.
- Stability: /metrics and protected routes 500 under Bun/vitest; gRPC surface not started.
- Config: Requires DB/Redis/NATS even when using memory repos; JWKS cache TTL hardcoded rather than using config value.

## Testing & Verification

- Current automated coverage is unstable (see Reality Check); integration tests for routes/metrics are failing.
- Metrics intended via /metrics (Prometheus); tracing to OTLP collector.
- Load and performance validation for SLOs should follow after functional fixes.

## Security Considerations

### Implemented (partial)
- JWKS verification util using jose with default 10 minute TTL, issuer/audience validation, and clock skew tolerance; not yet stable under tests.
- Scope-based authorization helper (`requireScopes`) and `optionalAuth` available; `requireAuth` populates request.user after JWKS verification.

### Planned / Missing
- Tenant isolation via requireTenantAccess (projects-service lookup with Redis cache).
- Secrets for S3/grid/NATS sourced from Vault; rate limiting per route/tenant using Redis.
- Artifact sanitization and access control (signed URLs).
