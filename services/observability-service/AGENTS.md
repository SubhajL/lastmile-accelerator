> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
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
