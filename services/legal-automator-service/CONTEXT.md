# legal-automator-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7503` and gRPC `50103` to deliver its function.

## Purpose & Scope
- Implements legal automator service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7503`
- gRPC: `50103` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=legal-automator-service
- SERVICE_PORT=7503
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=legal-automator-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-legal • Slack: #lma-legal-automator-service • PagerDuty: lma-legal-automator-service
