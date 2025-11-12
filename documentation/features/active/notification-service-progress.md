# Notification Service — Implementation Progress Tracker

Last Updated: 2025-11-12
Specification: ../active/notification-service-spec.md

## Overview
MVP targets event-driven notifications for Snapshot/Fix/Publish and reliability/observability baselines. Core flows and templates exist; production hardening, persistence, recipient resolution, and ops dashboards remain.

## Phase Completion Summary

| Phase          | Status | Completion | Notes |
|---------------:|:------:|:----------:|-------|
| P0 Sprint 1–2  |  ⏳    |   60%      | Core consumers/channels/templates in place; persistence/DLQ pending |
| P0 Sprint 3–4  |  ⏳    |   20%      | Recipient integration and dashboards TBD |
| P1 Sprint 5–6  |  ⏳    |   10%      | Prefs API/UI scaffolds exist; needs end‑to‑end |
| Hardening      |  ⏳    |    0%      | To be scheduled |

## Graphite Staged Diff Plan

Legend: Each diff is a small, single‑concern change (≤ ~200 LOC), stacked with Graphite. Provide PR titles and acceptance criteria for quick review. Owners reflect primary drivers; external owners are dependencies only.

1) SD‑01: Centralize NATS subjects and bindings
   - Branch: chore/nt-subjects-registry
   - Scope: Add constants/types for subjects; wire router to use them; no behavior change.
   - Tests: Update/add unit tests asserting subject names used by consumers.
   - Owner: Notifications. Depends on: —

2) SD‑02: Channel reliability defaults (timeouts/retries/circuit breaker)
   - Branch: feat/channel-reliability-defaults
   - Scope: Configure safe defaults; make env‑driven; no outbox yet.
   - Tests: Channel unit tests with timeouts/retry counts; CB trips/recovery.
   - Owner: Notifications. Depends on: SD‑01.

3) SD‑03: Outbox schema/types (inert)
   - Branch: feat/outbox-schema
   - Scope: Add PG schema DDL and TS types; do not call from dispatcher yet.
   - Tests: Migration/DDL snapshot; type compile.
   - Owner: Notifications. Depends on: —

4) SD‑04: Idempotent enqueue using outbox (feature‑flagged)
   - Branch: feat/dispatcher-outbox-idempotency
   - Scope: Write idempotency key + message to outbox; dedupe on enqueue; flag to enable.
   - Tests: Dispatcher unit tests covering duplicate suppression and retries.
   - Owner: Notifications. Depends on: SD‑03.

5) SD‑05: DLQ subject + redelivery policy
   - Branch: feat/dlq-and-redelivery
   - Scope: Publish to DLQ on poison errors; simple redelivery backoff util.
   - Tests: Router/consumer tests push failing payloads to DLQ.
   - Owner: Platform (dep), Notifications (impl). Depends on: SD‑01.

6) SD‑06: Admin send‑test endpoint (auth‑guarded)
   - Branch: feat/admin-send-test-endpoint
   - Scope: Minimal POST /admin/send‑test; requires JWT role; uses dispatcher.
   - Tests: App route tests with auth mocks.
   - Owner: DevEx (dep), Notifications (impl). Depends on: SD‑02.

7) SD‑07: Provider secrets via Vault wiring
   - Branch: feat/provider-secrets-vault
   - Scope: Load email/webhook/slack creds from Vault; helm value examples.
   - Tests: Config unit tests; mocks for Vault client.
   - Owner: Platform (dep), Notifications (impl). Depends on: —

8) SD‑08: Helm probes/resources and env/secret refs
   - Branch: chore/helm-probes-and-resources
   - Scope: Liveness/readiness probes; CPU/mem; secret/env refs.
   - Tests: Helm template lint/check (if available) or review checklist.
   - Owner: Platform. Depends on: SD‑07.

9) SD‑09: Recipient resolver adapter + FF mock
   - Branch: feat/recipient-resolver-adapter
   - Scope: Define interface; add FF‑based mock; no external call yet.
   - Tests: Consumer tests using mock resolver.
   - Owner: Notifications. Depends on: SD‑01, SD‑04.

10) SD‑10: Integrate projects‑service resolver
   - Branch: feat/recipient-resolver-projects
   - Scope: Wire HTTP/gRPC client; feature flag to switch from mock.
   - Tests: Adapter unit tests; consumer integration test with stubbed client.
   - Owner: Notifications; External dep: Projects. Depends on: SD‑09.

11) SD‑11: Expose delivery metrics for dashboards
   - Branch: feat/metrics-delivery-counters-hist
   - Scope: Prom/OTel metrics: delivered_total, retries_total, dlq_total, delivery_latency_ms.
   - Tests: Metrics unit tests; scrape text contains new series.
   - Owner: Notifications; Observability consumes. Depends on: SD‑04, SD‑05.

12) SD‑12: SLO definitions and alert thresholds (docs/config)
   - Branch: docs/slo-notifications
   - Scope: Define target SLOs and alert thresholds (p95 latency, success rate, DLQ growth) and names/labels.
   - Tests: — (reviewed config/docs).
   - Owner: Observability. Depends on: SD‑11.

Checklist (trackable via Graphite stack):
- [x] SD‑01  - NATS subjects centralized
- [ ] SD‑02  - Channel reliability defaults
- [ ] SD‑03  - Outbox schema/types
- [ ] SD‑04  - Idempotent enqueue (flagged)
- [ ] SD‑05  - DLQ + redelivery
- [ ] SD‑06  - Admin send‑test endpoint
- [ ] SD‑07  - Vault provider secrets wiring
- [ ] SD‑08  - Helm probes/resources
- [ ] SD‑09  - Recipient resolver adapter + FF mock
- [ ] SD‑10  - Projects resolver integration
- [ ] SD‑11  - Delivery metrics exposed
- [ ] SD‑12  - SLOs and alerts documented

## What needs to be done next
Use Graphite stack order:
1) SD‑01 → 2) SD‑02 → 3) SD‑03 → 4) SD‑04 → 5) SD‑05, then in parallel: SD‑06, SD‑07, SD‑08 → 9) SD‑09 → 10) SD‑10 → 11) SD‑11 → 12) SD‑12.
This preserves small, reviewable diffs while unblocking dashboards (metrics) only after idempotency/DLQ are in place.

## Blockers/Issues
- Recipient directory dependency (projects‑service) availability and API contract.
- Vault/secret paths for provider credentials need finalization.
- NATS subject naming and multi‑env configuration standardization.
