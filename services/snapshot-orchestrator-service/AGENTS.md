> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/snapshot-orchestrator-service-spec.md
- Progress:    ../../documentation/features/active/snapshot-orchestrator-service-progress.md
- Planned:     ../../documentation/features/planned/snapshot-orchestrator-service-spec.md

## Package Identity

- Go service orchestrating snapshots (REST 7054).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/snapshot-orchestrator-service/main.go`; implementation under `internal/**`.

## Touch Points

- `cmd/snapshot-orchestrator-service/main.go`, `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
