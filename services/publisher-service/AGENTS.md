> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/publisher-service-spec.md
- Progress:    ../../documentation/features/active/publisher-service-progress.md
- Planned:     ../../documentation/features/planned/publisher-service-spec.md

## Package Identity

- Go service publishing events/content (REST 7201).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/publisher-service/main.go`; domain code under `internal/**`.
- Tests are `*_test.go` colocated.

## Touch Points

- `cmd/publisher-service/main.go`, `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
