> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/perf-coach-service-spec.md
- Progress:    ../../documentation/features/active/perf-coach-service-progress.md
- Planned:     ../../documentation/features/planned/perf-coach-service-spec.md

## Package Identity

- Go service for performance coaching (REST 7302).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Implementation under `internal/**` (handlers, services, repository).

## Touch Points

- `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
