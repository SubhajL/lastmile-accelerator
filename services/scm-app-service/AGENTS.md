> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/scm-app-service-spec.md
- Progress:    ../../documentation/features/active/scm-app-service-progress.md
- Planned:     ../../documentation/features/planned/scm-app-service-spec.md

## Package Identity

- Go SCM app integration (REST 7051).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/scm-app-service/main.go`; domain code under `internal/**`.

## Touch Points

- `cmd/scm-app-service/main.go`, `internal/**`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build
```
