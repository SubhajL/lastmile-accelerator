# perf-coach-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7302` and gRPC `50082` to deliver its function.

## Purpose & Scope
- Implements perf coach service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7302`
- gRPC: `50082` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=perf-coach-service
- SERVICE_PORT=7302
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=perf-coach-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-perf • Slack: #lma-perf-coach-service • PagerDuty: lma-perf-coach-service
