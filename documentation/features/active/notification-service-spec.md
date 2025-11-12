# Notification Service — Technical Specification (MVP)

Document Name: Notification Service Implementation Plan
Date: 2025-11-12
Version: v0.1 (Active)
Status: Active

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
