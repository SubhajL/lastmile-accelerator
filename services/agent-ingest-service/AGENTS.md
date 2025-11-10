> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/agent-ingest-service-spec.md
- Progress:    ../../documentation/features/active/agent-ingest-service-progress.md
- Planned:     ../../documentation/features/planned/agent-ingest-service-spec.md

## Package Identity

- Go microservice for agent ingest (REST 7053, gRPC 50043; see service_catalog.yaml).

## Setup & Run

```bash
make build
make test
make run
```

## Patterns & Conventions

- Entry: `cmd/agent-ingest-service/main.go`.
- Application code under `internal/**`; tests are `*_test.go` colocated.

## Touch Points

- `cmd/agent-ingest-service/main.go`
- `internal/**`
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'func Test' internal || true
grep -R -nE 'http\.(Handle|ListenAndServe)' internal || true
```

## Pre-PR Checks

```bash
make test && make build
```
