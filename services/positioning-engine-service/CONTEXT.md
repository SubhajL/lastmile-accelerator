# positioning-engine-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7501` and gRPC `50101` to deliver its function.

## Purpose & Scope
- Implements positioning engine service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7501`
- gRPC: `50101` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=positioning-engine-service
- SERVICE_PORT=7501
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=positioning-engine-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-position • Slack: #lma-positioning-engine-service • PagerDuty: lma-positioning-engine-service
