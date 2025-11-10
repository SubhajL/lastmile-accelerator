> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/secrets-env-service-spec.md
- Progress:    ../../documentation/features/active/secrets-env-service-progress.md
- Planned:     ../../documentation/features/planned/secrets-env-service-spec.md

## Package Identity

- Go service for env secrets management (REST 7104).

## Setup & Run

```bash
make build     # builds ./bin/app
make test      # go test ./...
make run       # SERVICE_PORT=7104 ./bin/app
```

## Patterns & Conventions

- Entry: `cmd/secrets-env-service/main.go`.
- Handlers: `internal/handlers/**` (see tests like `internal/handlers/response_test.go`).
- Errors/retries: `internal/errors/**` with tests.
- Security: `internal/security/**` (e.g., rate limit token bucket).
- Repositories: `internal/repository/**`.

## Touch Points

- `cmd/secrets-env-service/main.go`
- `internal/handlers/response.go` (+ tests)
- `internal/errors/*` (+ tests)
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'func .*Handler\(' internal/handlers
grep -R -nE 'func Test' internal
```

## Pre-PR Checks

```bash
make test && make build
```
