# Notification Service — Implementation Progress Tracker

Last Updated: 2025-11-21
Specification: ../active/notification-service-spec.md

## Overview
The service has foundations (Redis queue with optional outbox dedup, channels with retry/timeout/CB, templates, preference routing, NATS subject registry), and **SD-01 Worker NATS Pump is now complete** - the worker actively consumes NATS events and processes notifications. The HTTP server (`src/app.ts`) only exposes `/healthz` and a stub `/metrics`; admin routes not yet mounted. No notification persistence/logging is performed and Vault/recipient integrations are absent.
**Current MVP Completion: ~40%** (building blocks exist + worker consumes NATS → enqueues → dispatches)

## Phase Completion Summary

| Phase          | Status | Completion | Notes |
|---------------:|:------:|:----------:|-------|
| P0 Sprint 1–2  |  ⏳    |   45%      | ✅ SD-01 complete: Worker consumes NATS, enqueues, dispatches. Admin route not mounted; no DB logging/observability wiring |
| P0 Sprint 3–4  |  ⏳    |   15%      | Quiet-hours deferral logic present; no digests, no recipient integration with projects-service, no delivery dashboards |
| P1 Sprint 5–6  |  ⏳    |   10%      | Preference store exists but no API/UI; SMS optional wiring only |
| Hardening      |  ⏳    |    0%      | No SLO/alert docs; CI gates allow failures; no DR runbook |

## Graphite Staged Diff Plan (updated)

Legend: small, single-concern diffs (≤ ~200 LOC). Ordered for unblockers → observability.

1) ✅ SD-01: Start worker NATS pump (COMPLETED 2025-11-21)
   - Branch: feat/admin-send-test-endpoint
   - Scope: Added `startNatsSubscriptions()` with lifecycle management, `startDispatcherLoop()` wrapper, worker mode flag, npm scripts
   - Implementation: Worker actively consumes 6 NATS subjects (snapshot.*, fixes.*, publish.*, checks.*, slo.*, errors.*), routes to domain consumers, enqueues notification jobs, dispatcher polls queue and dispatches to channels
   - Quality: All 148 tests passing, proper error handling, graceful shutdown, type-safe contracts, QCHECK reviewed and approved
   - Tests: Updated all worker tests with proper mocks for async lifecycle

2) SD-02: HTTP surface + admin send-test
   - Branch: feat/http-admin-endpoints
   - Scope: register Fastify routes (incl. `src/routes/admin.send-test.ts`) with auth guard; expose health/metrics from prom client; share queue handle.
   - Tests: route tests with JWT mock; metrics snapshot contains delivery counters.

3) SD-03: Delivery persistence & outbox lifecycle
   - Branch: feat/delivery-persistence
   - Scope: write `notification_logs`/`delivery_attempts`/`notification_outbox_messages` on ack/nack; ensure dedup uses DB and status transitions.
   - Tests: queue/dispatcher integration covering dedup + persisted status changes.

4) SD-04: DLQ + redelivery
   - Branch: feat/dlq-nats-redelivery
   - Scope: publish DLQ subject (e.g., `notifications.dlq`), add admin retry endpoint, keep Redis DLQ as store with backoff.
   - Tests: dispatcher/queue tests push poison jobs to DLQ and retry via admin endpoint.

5) SD-05: Metrics & alerts
   - Branch: feat/metrics-prom
   - Scope: Prom counters/histogram (`notification_sent_total`, `notification_retry_total`, `notification_dlq_total`, latency), `/metrics` endpoint, alert rules doc.
   - Tests: metrics unit tests asserting emitted series + histogram buckets.

6) SD-06: Provider secrets via Vault
   - Branch: feat/vault-secret-wiring
   - Scope: optional `VAULT_ENABLED`; hydrate SMTP/Twilio/Slack/Webhook creds from Vault with env fallback; helm examples.
   - Tests: config + vault client mocks verifying fallback and required paths.

7) SD-07: Recipient resolver integration
   - Branch: feat/recipient-resolver-projects
   - Scope: feature-flagged projects-service client (HTTP/gRPC) with mock fallback; resolver interface used by consumers.
   - Tests: resolver and consumer integration with stubbed client + flag switching.

8) SD-08: SLOs and alert thresholds
   - Branch: docs/slo-notification-service
   - Scope: Document SLO targets (success rate, p95 latency, DLQ growth) tied to metric names; add alert thresholds.
   - Tests: — (docs review)

Checklist:
- [ ] SD-01  - Worker NATS pump wired
- [ ] SD-02  - HTTP/admin routes exposed with metrics
- [ ] SD-03  - Delivery persistence + outbox lifecycle
- [ ] SD-04  - DLQ subject + redelivery
- [ ] SD-05  - Metrics counters/histogram
- [ ] SD-06  - Vault provider secrets
- [ ] SD-07  - Recipient resolver integration
- [ ] SD-08  - SLOs + alerts documented

## What needs to be done next
1) Wire the worker to actually consume NATS subjects and start the dispatcher (SD-01), otherwise no notifications flow.  
2) Mount HTTP/admin routes and real Prom metrics (SD-02, SD-05) so operators can trigger/test and monitor delivery.  
3) Add persistence + DLQ/redelivery (SD-03, SD-04) to meet idempotency and failure-handling requirements.  
4) Finish Vault + recipient integrations (SD-06, SD-07) and capture SLOs (SD-08).

## Blockers/Issues
- `loadConfig` requires Vault/SMTP env vars even for tests; hard to run locally without stubbing (`VAULT_ADDR`, etc.).
- No runnable path ties the worker to the HTTP server or start scripts; `pnpm start` only runs `/healthz`/stub `/metrics`.
- Sparse checkout excludes `.github` directory locally; CI cannot be inspected without `git show` and is currently weak (see CI section).

## Detailed Implementation Status (2025-11-21 Analysis)

### ✅ Foundations present
- NATS subject registry + router and consumers for snapshot/fixes/publish/checks/slo/errors (`src/events/subjects.ts`, `src/consumers/*`).
- Redis priority queue with backoff + optional PG outbox dedup (`src/notifications/queue.ts`, `src/outbox/pg-outbox.ts`).
- Channels with timeout/retry/circuit breaker and optional fallback (email, webhook, slack, SMS skeleton, in-app) (`src/channels/**`).
- Template store/loader/validator with handlebars bundles for core events (`src/templates/**`).
- Preference store + routing engine for quiet-hours/defer/block decisions (`src/prefs/store.ts`, `src/routing/engine.ts`).
- OTel tracing initialized; spans in dispatcher/channels (`src/telemetry.ts`, `src/telemetry/tracing.ts`).

### ⚠️ Gaps to reach MVP
- Worker never starts NATS subscriptions (`startOnce` unused) and run loop only processes an empty queue (`src/worker.ts`, `src/bootstrap/runtime.ts`).
- Fastify app exposes only `/healthz` and a stub `/metrics`; admin send-test route exists but is never registered (`src/app.ts`, `src/routes/admin.send-test.ts`).
- No persistence of notification logs/delivery attempts; DB client only used for outbox dedup.
- Metrics are placeholders; no Prom counters/histograms or alerting hooks (`src/metrics/otel.ts` only increments via meter).
- Vault creds parsed but not consumed anywhere; service fails to boot without Vault env vars even though secrets are unused (`src/config.ts`).
- Recipient resolution relies solely on payload/userId; no projects-service integration or feature flag (`src/recipients/resolvers.ts`).
- Quiet-hours defer decisions exist but no digest scheduler to honor deferrals; no DLQ redelivery API.

### ❌ Not implemented
- SLO/alert documentation; DR/rollback runbook.
- End-to-end test that covers NATS → queue → dispatcher → channel.
- CI gates that fail on typecheck/lint/test for this service.

## CI/CD Gates Status
- **PR CI:** `.github/workflows/pr-ci.yml` runs only db-guardian-service Go tests; notification-service code is not validated on PRs.  
- **Push-to-main:** `.github/workflows/ci-notification-service.yml` calls `_service-ci.yml`, but node steps run `bun run typecheck` and `bun run test` with `|| true`, so failures do not fail the job. Hadolint/gitleaks/OPA run but do not cover service correctness.  
- **Net:** No gate exercises the worker start path or metrics route; image build runs on push to `main` only.

## Critical Path to MVP Completion
1) Make the service actually consume events and expose admin/metrics endpoints (SD-01, SD-02).  
2) Add persistence + metrics + DLQ/redelivery to satisfy reliability/observability requirements (SD-03, SD-04, SD-05).  
3) Integrate Vault + recipient resolver + document SLO/alerts (SD-06, SD-07, SD-08).
