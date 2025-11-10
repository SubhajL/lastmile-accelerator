# webhook-relay-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7903` and gRPC `50123` to deliver its function.

## Purpose & Scope
- Implements webhook relay service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7903`
- gRPC: `50123` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=webhook-relay-service
- SERVICE_PORT=7903
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=webhook-relay-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-webhook • Slack: #lma-webhook-relay-service • PagerDuty: lma-webhook-relay-service
