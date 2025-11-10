# db-guardian-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7105` and gRPC `50065` to deliver its function.

## Purpose & Scope
- Implements db guardian service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7105`
- gRPC: `50065` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=db-guardian-service
- SERVICE_PORT=7105
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=db-guardian-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-dbguard • Slack: #lma-db-guardian-service • PagerDuty: lma-db-guardian-service
