> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/authz-matrix-service-spec.md
- Progress:    ../../documentation/features/active/authz-matrix-service-progress.md
- Planned:     ../../documentation/features/planned/authz-matrix-service-spec.md

## Package Identity

- Node/TypeScript service for authorization matrix evaluation.

## Setup & Run

```bash
# From repo root
bun install
bunx turbo run build --filter=authz-matrix-service
make run
```

## Patterns & Conventions

- Entry: `src/index.ts`. Use colocated `*.test.ts` for new logic.

## Touch Points

- `src/index.ts`, `package.json`, `tsconfig.json`
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'fastify\.(get|post|put|delete)\(' src || true
```

## Pre-PR Checks

```bash
make test && make build
```
