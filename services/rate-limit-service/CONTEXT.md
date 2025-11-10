# rate-limit-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7204` and gRPC `50074` to deliver its function.

## Purpose & Scope
- Implements rate limit service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7204`
- gRPC: `50074` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=rate-limit-service
- SERVICE_PORT=7204
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=rate-limit-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-rls • Slack: #lma-rate-limit-service • PagerDuty: lma-rate-limit-service
