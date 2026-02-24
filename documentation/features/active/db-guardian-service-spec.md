# db-guardian-service Technical Specification

**Document Name:** DB Guardian Service Implementation Plan  
**Date:** 2025-11-21
**Version:** 0.1.1
**Status:** Active
**Completion:** 75% toward MVP (9 tasks remaining)

## Executive Summary
The db-guardian-service guards customer databases by validating migrations, enforcing least-privilege roles, and recommending indexes. It exposes REST and gRPC APIs, persists audits/recommendations/policies, and publishes events so Projects/Publisher/Fix services can gate deploys. Success for MVP means: accept a project-scoped request, analyze the target database (not the service metadata DB), store and surface findings, and emit events without leaking secrets.

## Architecture Overview
- HTTP server on `:7105` with `/healthz`, `/metrics`, `/api/*`, and `/v1/projects/{id}/db/*`; gRPC on `:50065`.
- Analyzer layer: `RoleAnalyzer`, `MigrationGuard`, `IndexAdvisor` built on `DBInspector`; Postgres implementation only.
- Persistence: Postgres schema for `db_connections`, `role_policies`, `migration_audits`, `index_recommendations`; access via repositories and services.
- Integrations: optional Redis (health), NATS events, Vault for secrets (planned), JWT/JWKS auth with optional simple dev mode.
- Observability: structured logger, OTel tracing, Prometheus HTTP metrics, health checks.
- Current gap: analyzers run against the single service DB connection; they do not yet resolve per-project DSNs from stored connections/Vault.

## Implementation Phases
- **Phase 1 – Foundations (Complete):** config, logging, telemetry, health, HTTP server, graceful shutdown.
- **Phase 2 – Data Layer (Complete):** migrations for connections/policies/audits/index recommendations, repository layer, transaction helper.
- **Phase 3 – Analyzer Orchestration (In Progress):** role/migration/index analyzers, event publishing, REST/gRPC endpoints. Missing: real SQL parsing, performance checks with live stats, project-scoped DB resolution, applied-index tracking.
- **Phase 4 – MVP Hardening (Planned):** fetch DSNs from Vault via project connections; run analyzers against target DB with `pg_stat_statements`; align REST/gRPC contracts to product spec; add auth scopes + rate limits; integration tests with Postgres; enforce CI gates (no `go test ... || true` on push); SLO/error budget dashboards.

## Testing & Verification
- **Current:** `go test ./...` green locally with 100% unit test coverage; unit tests for analyzers, server, auth, gRPC handlers, repositories, telemetry using sqlmock. Bench harness exists.
- **Runbook:** `make verify` (vet + unit tests), `make quality` (vet, lint, test, build), `make grpc-codegen` to keep stubs fresh.
- **Gaps for MVP:**
  - No integration tests with real Postgres (need Testcontainers with `pg_stat_statements`)
  - No contract tests for REST `/v1/projects/{id}/db/*` or gRPC endpoints
  - No E2E tests for project DB resolution flow (Vault → DSN → analysis)
  - CI tests allowed to fail (`|| true` in workflows)

## Security Considerations
- Use JWT/JWKS auth in non-dev; simple authenticator is dev-only. Enforce scopes per route.
- Store DSNs as Vault refs only; never log raw DSNs or SQL. Fetch per project before analysis.
- Run analyzers with least-priv DB roles; avoid superuser connections. Validate inputs to prevent injection in dynamic SQL.
- Ensure NATS/event payloads omit secrets and include traceparent. Prefer mTLS for gRPC/mesh and TLS-terminating proxy for HTTP.
