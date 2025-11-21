# Notification Service — Technical Specification (MVP)

Document Name: Notification Service Implementation Plan
Date: 2025-11-21
Version: v0.2 (Reality Check)
Status: Active — MVP not yet met (see Reality Check)

## Executive Summary
Provide event-driven notifications for LMA core flows (Snapshot, Fix List, Fixes Applied, Publish Safely/Undo, Checks, Errors, SLO) with reliable delivery, templating, retries, and observability. MVP focuses on email, webhook, and Slack delivery, wired to NATS events, with strict idempotency and basic rate/circuit protection.

## Architecture Overview
- Service: Fastify (REST 7902) with `/healthz` and `/metrics`.
- Ingress: NATS JSON subscriptions via `events/nats.ts` and `events/nats-subscribe.ts`; routing via `consumers/router.ts`.
- Domain Consumers: `snapshot-consumer`, `fixes-consumer`, `publish-consumer`, `errors-consumer`, `slo-consumer`.
- Queue/Dispatch: `notifications/queue`, `notifications/dispatcher` with retry/backoff and idempotent enqueue.
- Channels: email (nodemailer), webhook HTTP, Slack (webhook), in‑app pubsub; SMS (twilio) present with limiter/fallback.
- Templates: handlebars store and validators; templates exist for major event types.
- Storage: Postgres (templates/logs/prefs), Redis (transient state), Vault (provider secrets), NATS (events).
- Observability: OpenTelemetry traces/metrics, Prometheus `/metrics`.
- Deploy: Helm manifests under `helm/`.

## Implementation Phases
P0 Sprint 1–2 (Core)
- Finalize NATS subjects/bindings for: snapshot.ready, fixlist.created, fixes.applied, publish.started|healthy|failed|rolledback.
- Harden Email/Webhook/Slack paths (timeouts, retries, circuit breaker).
- Validate template parameters; add golden snapshots for core templates.
- Persist notification logs/outbox with idempotency keys; implement enqueue de‑dupe; DLQ subject policy.
- Minimal admin endpoints: send‑test; confirm health/metrics.

P0 Sprint 3–4 (Integrations/Obs)
- Integrate recipient resolution with projects‑service; feature‑flag fallback.
- Add basic batching/digests and quiet‑hours for noisy events.
- Extend consumers to dep‑governance/db‑guardian/test‑lab critical events; corresponding templates.
- Dashboards/alerts for delivery success rate, p95 latency, DLQ growth.

P1 Sprint 5–6 (Preferences/Channels)
- Preferences API/UI and routing rules per user/channel; legal/unsubscribe handling.
- Harden SMS/push; provider failover matrices; quota/billing hooks.
- In‑app inbox persistence and read receipts.

Hardening (Weeks 13–14)
- SLOs, HA, DR runbook, security pen test, chaos tests for provider outages.

## Testing & Verification
- Unit: consumers, router, queue/dispatcher, channels, templates (including validation and rendering).
- Integration: NATS subscription adapter, Redis/PG happy paths, webhook timeouts/retries.
- Golden snapshots: handlebars template outputs for key events.
- E2E (mocked providers): event → dispatch → channel deliver → log/idempotency assertion.

## Security Considerations
- PII minimization in payloads and logs; redact secrets.
- Vault-backed provider secrets; no credentials in images.
- Per‑tenant rate limits and circuit breakers; exponential backoff on retries.
- Idempotency keys and outbox to prevent duplicate notifications; DLQ for poison messages.
- JWT/tenant scoping for any admin/test endpoints; audit important actions.

---

## Implementation Reality Check (2025-11-21)

### ✅ Implemented foundations
- Subject registry + consumers for snapshot/fixes/publish/checks/slo/errors (`src/events/subjects.ts`, `src/consumers/*`).
- Redis priority queue with backoff and optional PG outbox dedup (`src/notifications/queue.ts`, `src/outbox/pg-outbox.ts`).
- Channels with timeout/retry/circuit breaker (email, webhook, slack) plus SMS skeleton and in-app publisher (`src/channels/**`).
- Template store/loader/validator with handlebars bundles for core events (`src/templates/**`).
- Preference store + routing engine for quiet-hours/defer/block decisions (`src/prefs/store.ts`, `src/routing/engine.ts`).
- OTel tracing setup and spans in dispatcher/channels (`src/telemetry.ts`, `src/telemetry/tracing.ts`).

### ⚠️ Missing or incomplete vs. spec
- Worker bootstrap does not start NATS subscriptions or dispatcher loop; service start script only runs `/healthz`/stub `/metrics` (`src/worker.ts`, `src/bootstrap/runtime.ts`, `src/index.ts`).
- Admin send-test route exists but is never registered; no API for preferences or DLQ redelivery (`src/routes/admin.send-test.ts`, `src/app.ts`).
- No persistence of `notification_logs`/`delivery_attempts`; DB is only used for outbox dedup.
- Observability is limited to tracing; no Prom metrics/counters/histograms and `/metrics` returns a placeholder (`src/app.ts`, `src/metrics/otel.ts`).
- Vault variables are required by config but not used to fetch provider secrets; service cannot boot without Vault env even though secrets remain env-based (`src/config.ts`).
- Recipient resolution relies solely on payload/userId; no projects-service integration or feature-flagged resolver (`src/recipients/resolvers.ts`).
- Quiet-hours deferral exists, but no digest scheduler or DLQ redelivery policy; NATS DLQ subject not emitted.

### 🎯 MVP readiness snapshot (vs. spec)
- Core functionality: ~30% (building blocks present; no event→deliver path running)
- Reliability patterns: ~30% (channel-level protections without DLQ/redelivery/persistence)
- Observability: ~15% (tracing only; metrics/alerts absent)
- Production readiness: ~10% (Vault integration, SLOs/alerts, CI gates outstanding)

**Overall MVP Completion: ~30%** — end-to-end notification flow is not yet runnable; see Graphite plan in `notification-service-progress.md` for steps to close the gaps.
