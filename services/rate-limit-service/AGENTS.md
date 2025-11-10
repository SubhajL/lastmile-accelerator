> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/rate-limit-service-spec.md
- Progress:    ../../documentation/features/active/rate-limit-service-progress.md
- Planned:     ../../documentation/features/planned/rate-limit-service-spec.md

## Package Identity

- Go rate limiting service (REST 7204).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Core logic under `internal/**` (e.g., token bucket limiter).
- Tests: `*_test.go` colocated.

## Touch Points

- `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
