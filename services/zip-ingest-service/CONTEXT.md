# zip-ingest-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7052` and gRPC `50042` to deliver its function.

## Purpose & Scope
- Implements zip ingest service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7052`
- gRPC: `50042` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=zip-ingest-service
- SERVICE_PORT=7052
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=zip-ingest-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-zip • Slack: #lma-zip-ingest-service • PagerDuty: lma-zip-ingest-service
