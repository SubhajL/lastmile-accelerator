# fix-engine-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7151` and gRPC `50067` to deliver its function.

## Purpose & Scope
- Implements fix engine service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7151`
- gRPC: `50067` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=fix-engine-service
- SERVICE_PORT=7151
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=fix-engine-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-fix • Slack: #lma-fix-engine-service • PagerDuty: lma-fix-engine-service
