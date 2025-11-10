# scaffold-secure-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7103` and gRPC `—` to deliver its function.

## Purpose & Scope
- Implements scaffold secure service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7103`
- gRPC: `—` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=scaffold-secure-service
- SERVICE_PORT=7103
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=scaffold-secure-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-scaffold • Slack: #lma-scaffold-secure-service • PagerDuty: lma-scaffold-secure-service
