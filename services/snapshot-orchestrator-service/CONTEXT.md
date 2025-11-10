# snapshot-orchestrator-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7054` and gRPC `50044` to deliver its function.

## Purpose & Scope
- Implements snapshot orchestrator service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7054`
- gRPC: `50044` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=snapshot-orchestrator-service
- SERVICE_PORT=7054
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=snapshot-orchestrator-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-snapshot • Slack: #lma-snapshot-orchestrator-service • PagerDuty: lma-snapshot-orchestrator-service
