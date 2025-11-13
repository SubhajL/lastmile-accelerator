# Secrets Env Service Foundations — Technical Specification

Document Name: secrets-env-service-foundations-spec
Date: 2025-11-12
Version: v1
Status: Complete

## Executive Summary
Initial delivery of secrets-env-service with HTTP+gRPC APIs, storage/integration layers, observability/security basics, and comprehensive tests.

## Architecture Overview
- HTTP API with middleware: panic recovery, request logging, JWT auth, scope enforcement, tenant isolation, tracecontext.
- gRPC server with interceptors: recovery, logging, auth/scopes, tracecontext.
- Data: Postgres migrations/repos, Redis metadata cache, audit log repo.
- Events: NATS publisher with envelope and traceparent propagation.
- Observability: OTel tracing (minimal provider), Prometheus /metrics endpoint.
- Security: mTLS optional, input validation, RBAC header augmentation, rate limiting (in-memory token bucket).

## Implementation Phases (Completed)
1–6 Foundations: domain models, config, logger, errors, Vault test client; Postgres scaffolding; unit tests.
7–9 HTTP layer: handlers (secrets, env parity, leak scan); middleware; router/server; auth/scope tests.
10 gRPC service: proto, generated server/client; bufconn tests for auth/scope and methods.
11–13 Data & Events: audit repo; Redis cache; NATS publisher; traceparent propagation with integration tests.
14–16 Observability & Security: OTel tracing/minimal; /metrics; mTLS; rate limiting; validation; RBAC augmentation; tests for rate limit, validation, RBAC, spans.

## Testing & Verification
All unit/integration tests for HTTP/gRPC/auth/scopes, cache/events, and observability/security pass.

## Security Considerations
RBAC via scopes/headers; optional mTLS; careful validation and rate limiting to mitigate abuse.
