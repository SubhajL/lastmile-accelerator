> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/billing-service-spec.md
- Progress:    ../../documentation/features/active/billing-service-progress.md
- Planned:     ../../documentation/features/planned/billing-service-spec.md

## Package Identity

- Go billing service (REST 7901).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry in `cmd/billing-service/main.go`; internals under `internal/**` if present.

## Touch Points

- `cmd/billing-service/main.go`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
