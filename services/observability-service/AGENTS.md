> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../CONTEXT.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/observability-service-spec.md
- Progress:    ../../documentation/features/active/observability-service-progress.md
- Planned:     ../../documentation/features/planned/observability-service-spec.md

## Package Identity

- Go service for SLOs, alerts, and health checks (REST 7301).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/observability-service/main.go`.
- Handlers: `internal/handlers/*` (e.g., `slo_handler.go`, `queries_handler.go`).
- Health: `internal/health/health.go` (+ tests).
- Repository/services: `internal/repository/*`, `internal/services/*`.
- Migrations: `migrations/**`.

## Touch Points

- `internal/handlers/slo_handler.go`
- `internal/health/health.go`
- `migrations/**`, Helm manifests under `helm/templates/`

## JIT Index Hints

```bash
grep -R -nE 'func .*Handler\(' internal/handlers
grep -R -nE 'func Test' internal
```

## Pre-PR Checks

```bash
make test && make build
```

## Environment Matrix

| Variable                      | Type     | Required | Default               | Example                               | Notes |
|------------------------------|----------|----------|-----------------------|---------------------------------------|-------|
| SERVICE_PORT                 | int      | yes      | -                     | 7301                                  | HTTP port |
| GRPC_PORT                    | int      | yes      | -                     | 50065                                 | gRPC port |
| DB_URL                       | url      | yes      | -                     | postgres://user:pass@db:5432/app      | Postgres DSN |
| REDIS_URL                    | url      | yes      | -                     | redis://redis:6379/0                  | Redis DSN |
| NATS_URL                     | url      | yes      | -                     | nats://nats:4222                      | NATS URL |
| OTEL_EXPORTER_OTLP_ENDPOINT  | url      | yes      | -                     | http://otel-collector:4317            | OTLP traces endpoint |
| JWKS_URL                     | url      | yes      | -                     | https://auth.example.com/jwks.json    | JWKS endpoint |
| JWT_ISSUER                   | string   | yes      | -                     | https://auth.example.com/             | JWT issuer |
| PROMETHEUS_URL               | url      | yes      | -                     | http://prometheus:9090                | Prometheus base URL |
| TEMPO_URL                    | url      | yes      | -                     | http://tempo:3200                     | Tempo base URL |
| LOKI_URL                     | url      | yes      | -                     | http://loki:3100                      | Loki base URL |
| VAULT_ADDR                   | url      | yes      | -                     | http://vault:8200                     | Vault address |
| VAULT_ROLE_ID                | string   | yes      | -                     | 00000000-0000-0000-0000-000000000000  | Vault AppRole |
| VAULT_SECRET_ID              | string   | yes      | -                     | 11111111-1111-1111-1111-111111111111  | Vault AppRole |
| ENV                          | string   | no       | dev                   | dev                                   | Environment name |
| SERVICE_NAME                 | string   | no       | observability-service | observability-service                 | Service identifier |
| JWT_AUDIENCE                 | string   | no       | -                     | api://observability                   | Optional JWT aud |
| SLO_EVAL_INTERVAL            | duration | no       | 1m                    | 1m, 5m                                | Scheduler interval |
| SLO_EVAL_WORKERS             | int      | no       | 4                     | 4                                     | Scheduler workers |

## Scopes Matrix

| RPC                 | Scope                 |
|---------------------|-----------------------|
| CreateSLO           | observability:write   |
| GetSLO              | observability:read    |
| ListSLOsByProject   | observability:read    |
| UpdateSLO           | observability:write   |
| DeleteSLO           | observability:write   |
| GetSLOStatus        | observability:read    |
| GetSLOHistory       | observability:read    |
| CreateAlert         | observability:write   |
| GetAlert            | observability:read    |
| ListAlertsBySLO     | observability:read    |
| UpdateAlert         | observability:write   |
| DeleteAlert         | observability:write   |
| GetAlertHistory     | observability:read    |
| IngestError         | observability:ingest  |
| ListErrorGroups     | observability:read    |
| GetErrorGroup       | observability:read    |
| ListGroupEvents     | observability:read    |
| ResolveGroup        | observability:read    |
| SearchTraces        | observability:read    |
| GetTrace            | observability:read    |
| SearchLogs          | observability:read    |
| Golden              | observability:read    |

## gRPC examples (grpcurl)

```bash
# Get SLO status
grpcurl -plaintext -H "authorization: Bearer <token>" \
  localhost:50065 lma.observability.v1.ObservabilityService.GetSLOStatus \
  '{"id":"slo-123"}'

# Search logs
grpcurl -plaintext -H "authorization: Bearer <token>" \
  localhost:50065 lma.observability.v1.ObservabilityService.SearchLogs \
  '{"project_id":"proj-1","q":"{app=\"api\"}","limit":10}'
```
