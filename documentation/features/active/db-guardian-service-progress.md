# db-guardian-service — Implementation Progress Tracker

**Last Updated:** 2025-11-21  
**Specification:** ./db-guardian-service-spec.md

## Overview
Foundations and data layer are shipped (config, telemetry, health, migrations, repositories). Core analyzers, REST, and gRPC endpoints exist and tests pass locally (100% unit test coverage), but analysis still runs against the single service DB connection, not project-scoped targets. **Status: 75% complete toward MVP.** Critical blockers: (1) project DB resolution via Vault, (2) CI gates broken (Go 1.24 not available, tests allowed to fail with `|| true`), (3) no integration tests with real Postgres, (4) heuristic-only migration parsing.

## Phase Completion Summary

| Phase                     | Status      | Completion | Notes                                                            |
|---------------------------|:-----------:|:----------:|------------------------------------------------------------------|
| Phase 1 – Foundations     | Done        | 100%       | Config, logging, telemetry, health, HTTP server.                 |
| Phase 2 – Data Layer      | Done        | 100%       | Schema migrations, repositories, transaction helper.             |
| Phase 3 – Analyzers & API | In Progress | ~60%       | Role/migration/index analyzers, REST/gRPC, events; lacks project DB resolution and deeper checks. |
| Phase 4 – MVP Hardening   | Planned     | ~10%       | Per-project Vault-backed DSNs, stronger parsing/perf checks, auth/rate limits, integration tests, CI enforcement. |

## Current Tasks
- [x] Ship foundations with health/metrics, config validation, telemetry.
- [x] Add schema, repositories, and transaction helper.
- [x] Wire baseline role/migration/index analyzers with REST and gRPC endpoints plus NATS publishing.
- [ ] Resolve target DB connections per project (Vault ref → DSN) and run analyzers against that DB.
- [ ] Improve migration parsing/performance signals (pg\_stat\_statements, table stats) and align API payloads to product spec.
- [ ] Harden CI (remove `go test ... || true` on push, pin Go to available version) and add Postgres-backed integration tests.
- [ ] Add auth scope mapping and rate limits to `/v1/projects/{id}/db/*`; audit event payloads for PII/secrets.

## What needs to be done next
1) Implement project-scoped DB resolution (lookup default connection → Vault → dedicated inspector) and run analyzers on that target.  
2) Replace heuristic SQL parsing with structured checks and perf signals (pg\_stat\_statements thresholds, lock risk by size), persisting applied/declined recommendations.  
3) Align REST/gRPC contracts to the MVP spec (incl. PRD topics/fields), and add contract/integration tests for `/v1/projects/{id}/db/*`.  
4) Tighten CI gates (push jobs must fail on tests; use available Go version) and add Postgres fixture in CI.  
5) Add auth scopes + basic rate limits; ensure health/metrics cover dependencies and events are trace-linked.

## Blockers/Issues

### P0 - Critical Blockers (Must fix to reach MVP)
1. **CI Broken (Go Version):** `.github/workflows/_service-ci.yml:21` and `ci-db-guardian-service.yml:18,45` use Go `1.24` (not available); must change to `1.23` or `1.22`
2. **CI Broken (Test Failures Ignored):** `_service-ci.yml:58` has `|| true` allowing tests to pass even when failing; must remove to enforce quality gates
3. **No Project DB Resolution:** Analyzers run against service metadata DB only; missing `ResolveProjectInspector` to fetch Vault DSN and connect to customer databases
4. **No Integration Tests:** Only unit tests with mocks; need Testcontainers-based tests against real Postgres with `pg_stat_statements`

### P1 - MVP Gaps (Required for production-readiness)
- Migration parsing is heuristic regex; need pg_query_go for accurate SQL parsing
- Index advisor doesn't use pg_stat_statements thresholds; recommendations not prioritized
- Events lack project metadata and don't sanitize SQL/DSN (PII/secret leakage risk)
- Auth defaults to permissive SimpleAuthenticator; JWT/JWKS not enforced
- Rate limit middleware exists but not wired to HTTP/gRPC servers
- No contract tests for `/v1/projects/{id}/db/*` REST or gRPC endpoints
