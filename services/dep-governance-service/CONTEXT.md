# dep-governance-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7106` and gRPC `50066` to deliver its function.

## Purpose & Scope
- Implements dep governance service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7106`
- gRPC: `50066` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=dep-governance-service
- SERVICE_PORT=7106
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=dep-governance-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-deps • Slack: #lma-dep-governance-service • PagerDuty: lma-dep-governance-service
