# ai-debugger-service — CONTEXT

## One-liner
See PRD. This service runs on REST `:7102` and gRPC `50062` to deliver its function.

## Purpose & Scope
- Implements ai debugger service responsibilities per PRD.
- Provides health endpoint `/healthz`.

## APIs & Ports
- REST: `:7102`
- gRPC: `50062` (if applicable)

## Dependencies
- Auth (OIDC/JWT), NATS, Postgres/Redis/S3 as needed.

## Env Vars
- ENV (dev|staging|prod)
- SERVICE_NAME=ai-debugger-service
- SERVICE_PORT=7102
- OTEL_EXPORTER_OTLP_ENDPOINT
- VAULT_ADDR / VAULT_ROLE_ID / VAULT_SECRET_ID

## SLOs/NFRs
- API p95 ≤ 150ms, Availability 99.9%

## Security
- mTLS in mesh + JWT scopes; no client secrets stored.

## Observability
- `/healthz` + `/metrics`; OTel spans tagged with `service.name=ai-debugger-service`

## Testing
- Unit tests for handler; integration stubs in CI.

## Ownership
- Team: @team-aidbg • Slack: #lma-ai-debugger-service • PagerDuty: lma-ai-debugger-service
