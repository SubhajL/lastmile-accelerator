# Test Lab Service — Implementation Progress Tracker

Last Updated: 2025-11-12

Specification: ./test-lab-service-spec.md

## Overview

Foundations, core REST APIs, repositories, observability, and runners are implemented with comprehensive tests. Event subscribers are wired behind a feature flag. gRPC surface and tenant/JWKS hardening remain outstanding. Makefile/Dockerfile need alignment with build output.

## Phase Completion Summary

| Phase                          | Status | Completion | Notes |
|-------------------------------:|:------:|:----------:|-------|
| Foundations (config/logger/otel/metrics) | ✅ | 100% | Config with Zod, Pino logger, OTel SDK, /metrics, /healthz |
| AuthN/Z (JWT + scopes)         | ⏳     | 70% | Auth plugin + requireScopes done; JWKS + optionalAuth + tenant enforcement pending |
| Data layer (PG + migrations)   | ✅     | 100% | Migrations + PG repos for scaffolds, test_runs, browser_test_runs, preview_envs |
| REST: scaffolds                | ✅     | 100% | CRUD with Zod validation and tests |
| REST: test runs                | ✅     | 100% | Create/list/get/update-status + tests |
| REST: browser runs             | ✅     | 100% | Create/list/get/update-status + tests |
| REST: previews                 | ✅     | 100% | Create/list/get/update/delete + tests |
| Runner (K8s Jobs + S3)         | ⏳     | 80% | Orchestrator and artifacts service implemented; production config hardening pending |
| Browser Grid (selenium)        | ⏳     | 70% | Runner implemented with basic flow; finalize suites and retries |
| Events (NATS)                  | ⏳     | 80% | Contracts + publishers + subscribers; wiring gated by flag |
| gRPC surface (50072)           | ❌     | 0%  | Not implemented |
| Platform (Dockerfile/Makefile) | ⏳     | 40% | Makefile points to ./bin/server; Dockerfile needs multi-stage build to dist |

## Completed Tasks / Phases

- Config module with Zod validation, getConfig/validateEnv.
- Logger with redaction, child context; global error handler.
- OpenTelemetry SDK + auto-instrumentations; Prometheus /metrics.
- Repos + migrations: scaffolds, test_runs, browser_test_runs, preview_envs.
- REST endpoints: scaffolds, test runs, browser runs, previews (with Zod schemas, scope guards).
- K8s runner and artifacts service; selenium grid runner; NATS publishers/subscribers (flagged).
- App wiring: createApp(), health/metrics, auth, routes, optional event wiring.

## MVP — Remaining Tasks (to move to Complete)

- [ ] Implement JWKS-based JWT verification with caching (replace shared secret usage in production). Owner: BE-Auth. Sprint: S1
- [ ] Add optionalAuth and requireTenantAccess hooks with project tenant lookup and Redis cache. Owner: BE-Auth. Sprint: S1
- [ ] Implement preview environment orchestration (K8s namespaces/services) and lifecycle management. Owner: Infra. Sprint: S2
- [ ] Implement gRPC server on 50072 for orchestration (CreateTestScaffold, ExecuteTests, GetTestResults, CreatePreview). Owner: BE-Core. Sprint: S3
- [ ] Enable and configure event subscribers in production (snapshot.ready → smoke runs, fixes.applied → regression). Owner: BE-Core. Sprint: S2
- [ ] Rate limiting and Redis caching for hotspots; define budgets per route. Owner: Platform. Sprint: S2
- [ ] Finalize Dockerfile (multi-stage build to dist) and fix Makefile run target. Owner: Platform. Sprint: S1
- [ ] SLO verification: add load tests and dashboards to ensure API p95 ≤ 150ms and 99.9% availability. Owner: SRE/Obs. Sprint: S3

## What needs to be done next

1) Security hardening: JWKS + tenant enforcement hooks.
2) Preview orchestration service and wiring to routes.
3) gRPC server and handlers mapped to existing services.
4) Enable event subscribers and verify end-to-end flows.
5) Platform fixes (Dockerfile/Makefile) and performance validation.

## Blockers/Issues

- JWKS endpoint and issuer/audience details required from auth provider.
- Kubernetes namespaces/permissions and container image for test runner must be provisioned.
- Selenium Grid URL and credentials for Safari/Firefox farms if not self-hosted.

## Owner & Sprint Breakdown

Legend Owners: BE-Auth (Auth/Scopes), BE-Core (API/Services), Infra (K8s/S3/NATS), Platform (Build/Runtime), SRE/Obs (SLO/Monitoring)

### Sprint S1
- [ ] JWKS verification with caching (Owner: BE-Auth)
- [ ] optionalAuth + requireTenantAccess + Redis cache (Owner: BE-Auth)
- [ ] Dockerfile (multi-stage) + Makefile run target fix (Owner: Platform)
- [ ] Bootstrap runs migrations on startup when REPO_BACKEND=pg (Owner: Platform)

### Sprint S2
- [ ] Preview env orchestration (K8s services/namespace lifecycle) (Owner: Infra)
- [ ] Enable event subscribers (snapshot.ready/fixes.applied) and E2E verification (Owner: BE-Core)
- [ ] Route-level rate limiting using Redis (Owner: Platform)

### Sprint S3
- [ ] gRPC server on 50072 (handlers wired to services) (Owner: BE-Core)
- [ ] Load tests + dashboards to validate p95 and availability (Owner: SRE/Obs)

## Graphite Stack Plan (staged diffs)

Use small, reviewable stacks (≤3 files, ≤250 LOC) with tests green at every step. Suggested branch names assume `gt` workflow; each bullet is a single stacked diff.

### S1-A — JWT Hardening (Owner: BE-Auth)
1. feat(config): add JWKS config keys + types; unit tests (files: src/config.ts, src/__tests__/unit/config.test.ts) — Branch: test-lab/jwks-config
2. feat(auth): add JWKS verifier util and tests (files: src/lib/jwks.ts, src/__tests__/unit/middleware/auth.test.ts) — Branch: test-lab/jwks-util
3. feat(auth): add optionalAuth hook + tests (files: src/middleware/auth.ts, tests) — Branch: test-lab/optional-auth
4. feat(auth): integrate registerAuthPlugin to use JWKS util behind env flag; update app wiring; tests (files: src/middleware/auth.ts, src/app.ts) — Branch: test-lab/auth-wire-jwks
5. chore(auth): default to JWKS in staging/prod, keep secret for tests only (files: src/config.ts) — Branch: test-lab/jwks-default

### S1-B — Tenant Access (Owner: BE-Auth)
1. feat(client): add projects client with Redis cache + tests (files: src/clients/projects.ts, tests) — Branch: test-lab/projects-client
2. feat(middleware): implement requireTenantAccess using client; unit tests — Branch: test-lab/tenant-mw
3. feat(routes): apply requireTenantAccess to scaffolds/runs/previews; integration tests — Branch: test-lab/tenant-wire
4. perf(auth): cache TTL tuning and error mapping; tests — Branch: test-lab/tenant-tuning

### S1-C — Platform (Owner: Platform)
1. chore(Dockerfile): multi-stage build to dist; HEALTHCHECK; expose 7202 — Branch: test-lab/docker
2. chore(Makefile): fix run target to `node dist/index.js`; add migrate target — Branch: test-lab/makefile
3. feat(app): auto-run migrations on startup when REPO_BACKEND=pg; tests — Branch: test-lab/auto-migrate

### S2-A — Preview Orchestration (Owner: Infra)
1. feat(k8s): preview manager client + service; unit tests — Branch: test-lab/preview-svc
2. feat(routes): wire preview service into routes; publish events; integration tests — Branch: test-lab/preview-wire
3. feat(preview): TTL extend + lifecycle states; tests — Branch: test-lab/preview-ttl

### S2-B — Events Enablement (Owner: BE-Core)
1. feat(config): feature flag for subscribers; defaults off — Branch: test-lab/events-flag
2. feat(app): enable subscribers in staging; close hooks; tests — Branch: test-lab/events-wire
3. test(e2e): inject NATS messages to verify snapshot.ready → smoke run — Branch: test-lab/events-e2e

### S2-C — Rate Limiting (Owner: Platform)
1. feat(rate): Redis-backed rate limiting middleware; unit tests — Branch: test-lab/rate-limit
2. chore(routes): apply per-route budgets; integration tests — Branch: test-lab/rate-apply

### S3 — gRPC & SLO (Owners: BE-Core, SRE/Obs)
1. feat(grpc): proto + server skeleton; unit tests — Branch: test-lab/grpc-proto
2. feat(grpc): handlers mapped to services; integration-light tests — Branch: test-lab/grpc-handlers
3. chore(slo): k6 load scripts + dashboards; acceptance thresholds — Branch: test-lab/slo-scripts

### Stretch / Parallel
- [ ] Tenant cache invalidation and TTL tuning (BE-Auth)
- [ ] Grid retries flake policy and screenshots/logs enrichment (BE-Core)
