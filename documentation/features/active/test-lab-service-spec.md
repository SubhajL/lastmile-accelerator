# Test Lab Service — Technical Specification

Document Name: test-lab-service Implementation Plan

Date: 2025-11-12

Version: 1.0

Status: Active

## Executive Summary

Test Lab Service provides automated test scaffolding, ephemeral preview environments, and cross-browser execution at scale. It exposes a REST API on port 7202 (Fastify/Node) and plans a gRPC surface on 50072. Core integrations: Postgres (persistence), Redis (cache/rate limit), NATS JetStream (events), S3 (artifacts), Kubernetes Jobs (isolated runs), Selenium Grid/BrowserStack (browser grid). Observability uses OpenTelemetry (traces) and Prometheus (metrics). SLOs: API p95 ≤ 150ms and 99.9% availability.

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

1) Foundations: config, logger, error handler, telemetry, metrics.
2) AuthN/Z: JWT auth + scope checks; tenant access enforcement (planned) and JWKS verification (planned).
3) Data: migrations + PG repositories for all core tables.
4) REST APIs: scaffolds, test runs, browser runs, previews.
5) Runners: K8s job orchestrator, artifacts service.
6) Browser Grid: selenium-based runner and retries.
7) Events: contracts, publishers, subscribers (feature-flagged wiring).
8) gRPC: orchestration surface (planned).

## Testing & Verification

- Unit and integration tests across config, middleware, repos, routes, services, and events (vitest).
- Metrics exposed via /metrics; use load tests to verify p95 ≤ 150ms.
- Tracing exported to OTLP collector for end-to-end latency and error analysis.

## Security Considerations

- JWT verification should use JWKS (planned) with caching and issuer/audience validation; current implementation uses fastify-jwt secret for tests.
- Scope-based authorization (requireScopes) enforced per-route; tenant isolation via requireTenantAccess planned.
- Secrets (S3, grid, NATS) must be sourced from Vault; no secrets in logs. Artifacts sanitized.
