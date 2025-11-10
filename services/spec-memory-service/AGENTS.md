> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/spec-memory-service-spec.md
- Progress:    ../../documentation/features/active/spec-memory-service-progress.md
- Planned:     ../../documentation/features/planned/spec-memory-service-spec.md

## Package Identity

- Go service for spec memory (REST 7101).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/spec-memory-service/main.go`; logic under `internal/**`.

## Touch Points

- `cmd/spec-memory-service/main.go`, `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
