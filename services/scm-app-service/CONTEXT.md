# scm-app-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7051` and gRPC `50041` to deliver its function.

## Purpose & Scope
- Implements scm app service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7051`
- gRPC: `50041` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=scm-app-service
- SERVICE_PORT=7051
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=scm-app-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-scm • Slack: #lma-scm-app-service • PagerDuty: lma-scm-app-service
