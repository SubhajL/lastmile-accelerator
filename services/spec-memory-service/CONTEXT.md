# spec-memory-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7101` and gRPC `50061` to deliver its function.

## Purpose & Scope
- Implements spec memory service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7101`
- gRPC: `50061` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=spec-memory-service
- SERVICE_PORT=7101
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=spec-memory-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-spec • Slack: #lma-spec-memory-service • PagerDuty: lma-spec-memory-service
