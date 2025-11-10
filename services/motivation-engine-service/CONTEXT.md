# motivation-engine-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7601` and gRPC `50111` to deliver its function.

## Purpose & Scope
- Implements motivation engine service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7601`
- gRPC: `50111` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=motivation-engine-service
- SERVICE_PORT=7601
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=motivation-engine-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-motivation • Slack: #lma-motivation-engine-service • PagerDuty: lma-motivation-engine-service
