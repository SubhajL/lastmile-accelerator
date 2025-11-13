> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/secrets-env-service-mvp-hardening-spec.md
- Progress:    ../../documentation/features/active/secrets-env-service-mvp-hardening-progress.md
- Planned:     ../../documentation/features/planned/secrets-env-service-enhancements-spec.md

## Graphite Staged‑Diff Plan (Stacked chain)

Legend: S1 = current sprint, S2 = next sprint. Each branch is stacked on the previous one (parent shown below).

| # | Branch                      | Parent                  | Scope (small, reviewable unit)                                                   | Files (primary)                               | gt create example                                                                                              | Owner  | Sprint |
|--:|-----------------------------|-------------------------|----------------------------------------------------------------------------------|-----------------------------------------------|------------------------------------------------------------------------------------------------------------------|:------:|:------:|
| 1 | svc/otel-config             | main                    | Add config fields: OTELEndpoint, OTELInsecure, OTELHeaders, OTELServiceName      | internal/config/config.go                     | `gt create svc/otel-config -am "feat(secrets-env-service): add OTEL config fields"`                             | SubhajL|  S1    |
| 2 | svc/otel-exporter           | svc/otel-config         | Implement OTLP exporter + TLS/insecure + headers; provider shutdown               | internal/observability/otel.go                | `gt create svc/otel-exporter -am "feat(secrets-env-service): add OTLP trace exporter and graceful shutdown"`    | SubhajL|  S1    |
| 3 | svc/main-otel-wiring        | svc/otel-exporter       | Init/shutdown wiring in main; pass config through                                 | cmd/secrets-env-service/main.go               | `gt create svc/main-otel-wiring -am "chore(secrets-env-service): wire OTEL exporter init/shutdown in main"`     | SubhajL|  S1    |
| 4 | svc/http-metrics            | svc/main-otel-wiring    | HTTP metrics: counters + histograms by method/route/status                        | internal/handlers/middleware.go (+ tests)     | `gt create svc/http-metrics -am "feat(secrets-env-service): add HTTP metrics counters and histograms"`          | SubhajL|  S1    |
| 5 | svc/grpc-metrics            | svc/http-metrics        | gRPC metrics: counters + latencies by method/code                                 | internal/grpc/server.go (+ tests)             | `gt create svc/grpc-metrics -am "feat(secrets-env-service): add gRPC metrics counters and duration histograms"` | SubhajL|  S1    |
| 6 | svc/domain-metrics          | svc/grpc-metrics        | Domain counters: secrets/parity/leak-scan increments                              | internal/service/* (+ tests)                  | `gt create svc/domain-metrics -am "feat(secrets-env-service): add domain metrics for secrets/parity/leak-scan"` | SubhajL|  S1    |
| 7 | svc/env-allowlist           | svc/domain-metrics      | Env allowlist validation + config default                                         | internal/handlers/secrets.go, config (+ tests)| `gt create svc/env-allowlist -am "feat(secrets-env-service): enforce env allowlist and defaults"`               | SubhajL|  S2    |
| 8 | svc/http-content-type       | svc/env-allowlist       | Enforce JSON Content-Type middleware                                              | internal/handlers/middleware.go (+ tests)     | `gt create svc/http-content-type -am "feat(secrets-env-service): enforce JSON Content-Type"`                    | SubhajL|  S2    |
| 9 | svc/http-body-limit         | svc/http-content-type   | Request body size limit via http.MaxBytesReader + HTTP_MAX_BODY_BYTES             | internal/handlers/middleware.go (+ tests)     | `gt create svc/http-body-limit -am "feat(secrets-env-service): add request body size limit"`                    | SubhajL|  S2    |
|10 | svc/rbac-claims             | svc/http-body-limit     | Prefer roles from JWT claims (HTTP+gRPC) with header override; add tests          | middleware + internal/grpc/server.go (+ tests)| `gt create svc/rbac-claims -am "feat(secrets-env-service): prefer roles from JWT claims; keep header override"` | SubhajL|  S2    |
|11 | svc/grpc-health-reflection  | svc/rbac-claims         | gRPC health service + reflection behind toggle                                    | internal/grpc/server.go (+ tests)             | `gt create svc/grpc-health-reflection -am "feat(secrets-env-service): enable gRPC health and reflection (toggle)"`| SubhajL|  S2    |
|12 | svc/startup-readiness       | svc/grpc-health-reflection| Startup readiness: PG fail-fast; Redis/NATS degraded; Vault health; smoke tests  | cmd/main.go (+ fakes/tests)                   | `gt create svc/startup-readiness -am "feat(secrets-env-service): add startup readiness checks"`                  | SubhajL|  S2    |
|13 | svc/docs-readme             | svc/startup-readiness   | Update service README: new env vars; curl/grpcurl examples                        | services/secrets-env-service/README.md        | `gt create svc/docs-readme -am "docs(secrets-env-service): document new env vars and examples"`                 | SubhajL|  S2    |

Notes:
- Workflow per slice: Make the minimal change set for the slice → run the shown `gt create <name> -am "message"` to both create the stacked branch and commit.
- Verify relationships with `gt branch info` (Parent field) and visualize with `gt ls`.
- Submit as a stack from the tip: `gt stack submit`; if trunk moves, `gt stack restack`.

## Package Identity

- Go service for env secrets management (REST 7104).

## Setup & Run

```bash
make build     # builds ./bin/app
make test      # go test ./...
make run       # SERVICE_PORT=7104 ./bin/app
```

## Patterns & Conventions

- Entry: `cmd/secrets-env-service/main.go`.
- Handlers: `internal/handlers/**` (see tests like `internal/handlers/response_test.go`).
- Errors/retries: `internal/errors/**` with tests.
- Security: `internal/security/**` (e.g., rate limit token bucket).
- Repositories: `internal/repository/**`.

## Touch Points

- `cmd/secrets-env-service/main.go`
- `internal/handlers/response.go` (+ tests)
- `internal/errors/*` (+ tests)
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'func .*Handler\(' internal/handlers
grep -R -nE 'func Test' internal
```

## Pre-PR Checks

```bash
make test && make build
```
