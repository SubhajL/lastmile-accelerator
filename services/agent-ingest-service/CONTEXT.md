# agent-ingest-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7053` and gRPC `50043` to deliver its function.

## Purpose & Scope
- Implements agent ingest service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7053`
- gRPC: `50043` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=agent-ingest-service
- SERVICE_PORT=7053
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=agent-ingest-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-agent • Slack: #lma-agent-ingest-service • PagerDuty: lma-agent-ingest-service
