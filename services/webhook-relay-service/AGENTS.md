> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/webhook-relay-service-spec.md
- Progress:    ../../documentation/features/active/webhook-relay-service-progress.md
- Planned:     ../../documentation/features/planned/webhook-relay-service-spec.md

## Package Identity

- Go service relaying webhooks (REST 7903).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Handlers under `internal/handlers/*`; repositories under `internal/repository/*`.
- Tests: `*_test.go` colocated.

## Touch Points

- `internal/handlers/**`, `internal/repository/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
