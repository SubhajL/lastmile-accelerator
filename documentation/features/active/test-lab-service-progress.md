# Test Lab Service — Implementation Progress Tracker

Last Updated: 2025-11-22

Specification: ./test-lab-service-spec.md

## Overview
- MVP not reached. The service has a Fastify skeleton (config, JWKS helper, repos, routes, telemetry stubs), but orchestration and stability gaps block usage.
- Latest local run `bun test` shows 45/143 failures driven by Fastify reply errors (ERR_HTTP_HEADERS_SENT on routes/metrics), missing auth surface (`authenticateRequest` removal), JWKS test fixtures failing jose verification, and Bun failing to resolve `putArtifact` in S3 helper for runner/browser-grid tests.
- CI workflows for PR and main branch have been added (2025-11-22), including quality gates, security scans, Docker build, SBOM generation, and GHCR push. gRPC surface is absent. Runner/grid/event flows are stubs and not connected to HTTP endpoints.
- Branch state: on `test-lab/add-github-actions-workflows` with new CI/CD workflows. Working tree has doc edits and .turbo cache artifacts.

## Phase Completion Summary
| Phase | Status | Notes |
| ---- | ---- | ---- |
| Foundations (config/logger/otel/metrics) | Partial | Zod config + logger + telemetry init exist; /metrics handler fails under Bun tests with ERR_HTTP_HEADERS_SENT. |
| AuthN/Z (JWT + scopes) | Partial | requireAuth/requireScopes implemented; registerAuthPlugin only decorates request.user; legacy `authenticateRequest` expected by tests is gone; JWKS util not using config cache TTL. |
| Data layer (PG + migrations) | Mostly | Migrations + pg-mem repos pass unit tests; memory backend only supports scaffolds. |
| REST: scaffolds/test runs/browser runs/previews | Partial/blocked | Routes defined, but integration tests 500 due to Fastify reply errors and auth mismatch. Test-run and preview routes register only with `REPO_BACKEND=pg`; no orchestration trigger. |
| Runner (K8s Jobs + S3) | Partial | Job manifest + orchestrator stub; placeholder command and no wiring from HTTP/events; S3 helper resolution failing under Bun. |
| Browser Grid | Partial | Runner class with screenshot upload stub; not integrated; tests blocked by S3 import error. |
| Events (NATS) | Partial | Contracts/publishers exist; subscriber wiring behind flag; wiring test currently errors on vi.mock usage. |
| gRPC surface (50072) | Not started | No proto/server implementation. |
| Platform (Dockerfile/Makefile) | Partial | Dockerfile unverified; Makefile `run` target points to missing `./bin/server`; CI workflows added (PR: quality+security, Main: Docker+SBOM+CVE scan+GHCR). |

## Current Implementation Snapshot
- App wiring: `src/app.ts` sets up telemetry, error handler, JWKS auth hooks, repos, `/healthz` + `/metrics`, REST routes; event subscribers gated by `ENABLE_EVENT_SUBSCRIBERS`.
- Auth: `requireAuth`/`optionalAuth` call `verifyJwt`; `authenticateRequest` removed but tests/docs still reference it.
- Data: PG migrations for scaffolds/test runs/browser runs/previews; pg-mem support for tests.
- Execution: Runner/browser-grid classes exist with stubbed behavior; no linkage from REST routes to orchestrators or to NATS publishers.
- Observability: telemetry init + metrics helper exist but currently fail under test harness.
- CI/CD: GitHub Actions workflows added - PR workflow (quality gates + security scans) and Main workflow (Docker build, SBOM generation, CVE scanning, GHCR push); Dockerfile/Makefile not validated.

## Build & Test Status (local)
- `bun run typecheck` ✅ (passes).
- `bun test` ❌ (45/143 failing). Failures: Fastify reply errors on REST + /metrics; auth plugin API drift (`authenticateRequest` expectations); JWKS jose verification with fixtures; S3 `putArtifact` resolution under Bun; telemetry shutdown expectation.
- Lint/build not run in this check.

## Blocking Issues (toward MVP)
1. Fastify replies double-send/ERR_HTTP_HEADERS_SENT on route and metrics handlers under Bun tests; need to debug error handler and route return behavior.
2. Auth surface drift: registerAuthPlugin only decorates `request.user`; legacy `authenticateRequest` and shared-secret path still in tests; align middleware with JWKS util and update tests/spec.
3. JWKS util tests failing because jose verify needs deterministic test JWKS plus config-driven cache TTL; optionalAuth path swallows errors differently than tests expect.
4. S3 helper import resolution (`putArtifact`) failing in Bun for runner/browser-grid tests; likely module format or mocking issue.
5. Orchestration not wired: test run creation does not enqueue runner/browser jobs or emit NATS events; preview orchestration absent; gRPC surface absent.
6. ~No CI workflows~ CI workflows added (2025-11-22); still need to validate Docker/Makefile.

## Next Steps (proposed)
1. Fix Fastify response handling for routes and `/metrics` so integration tests pass under Bun/vitest.
2. Decide and document supported auth middleware (`authenticateRequest` vs `requireAuth`/`optionalAuth`), update code/tests accordingly, and ensure JWKS cache uses config values.
3. Resolve Bun module resolution for S3 helper and adjust runner/browser-grid tests; add harness-friendly stubs for K8s/S3/NATS.
4. Wire test-run and browser-run routes to orchestrator/event publish stubs (feature-flagged) and add minimal happy-path integration tests.
5. ~Add CI workflows for PR/main~ ✅ CI workflows added (2025-11-22) with quality gates and security scans; still need to fix Makefile `run` target and validate Dockerfile build.
6. Leave gRPC/preview env lifecycle/rate limiting in backlog with explicit deferral in spec.

## CI/CD Implementation (2025-11-22)

### GitHub Actions Workflows Added
1. **PR Workflow** (`.github/workflows/test-lab-service-pr.yml`):
   - Triggers on PR to main/master/db-guardian-service branches
   - Runs quality gates: typecheck, lint, test (with Node runner), build
   - Runs security scans: gitleaks (secrets), hadolint (Dockerfile)
   - Uses Bun with Turborepo filtering

2. **Main Branch Workflow** (`.github/workflows/test-lab-service-main.yml`):
   - Triggers on push to main/db-guardian-service branches
   - Builds application with Turborepo
   - Builds Docker image with caching
   - Generates SBOM with Anchore
   - Scans for CVEs with Trivy (fails on CRITICAL/HIGH)
   - Pushes to GitHub Container Registry (ghcr.io)
   - Uploads SBOM as artifact (90-day retention)

### Key Configurations
- Path filters include service code, shared packages, and workflow files
- Concurrency groups prevent duplicate runs
- Uses Node runner for tests to ensure compatibility
- Docker tags: `{branch}-{sha}` and `latest` for main branch
- Caching for Bun dependencies and Docker layers

## Legacy Planning Notes
Prior sprint breakdowns and Graphite stack plan in earlier revisions assumed auth hardening work was already green. Treat them as stale; refer to git history if needed for archival context.
