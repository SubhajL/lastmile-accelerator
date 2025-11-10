> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/db-guardian-service-spec.md
- Progress:    ../../documentation/features/active/db-guardian-service-progress.md
- Planned:     ../../documentation/features/planned/db-guardian-service-spec.md

## Package Identity

- Go service guarding DB migrations and performance.

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Core analyzers under `internal/analyzer/*.go` with thorough `*_test.go`.
- Config, database, cache, telemetry under `internal/{config,database,cache,telemetry}`.

## Touch Points

- `internal/analyzer/index_advisor.go`, `internal/analyzer/migration_guard.go`
- Tests: `internal/analyzer/*_test.go`

## JIT Index Hints

```bash
grep -R -nE 'func Test' internal
grep -R -nE 'migration_guard|index_advisor' internal/analyzer
```

## Pre-PR Checks

```bash
make test && make build
```
