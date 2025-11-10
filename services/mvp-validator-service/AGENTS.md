> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/mvp-validator-service-spec.md
- Progress:    ../../documentation/features/active/mvp-validator-service-progress.md
- Planned:     ../../documentation/features/planned/mvp-validator-service-spec.md

## Package Identity

- Node/TypeScript Fastify service for MVP validation flows.

## Setup & Run

```bash
pnpm i
pnpm --filter mvp-validator-service build
make run
```

## Patterns & Conventions

- Entry: `src/index.ts`. Co-locate tests as `*.test.ts` in `src/**` when adding features.

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
