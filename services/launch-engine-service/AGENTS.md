> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/launch-engine-service-spec.md
- Progress:    ../../documentation/features/active/launch-engine-service-progress.md
- Planned:     ../../documentation/features/planned/launch-engine-service-spec.md

## Package Identity

- Node/TypeScript service orchestrating launches.

## Setup & Run

```bash
pnpm i
pnpm --filter launch-engine-service build
make run
```

## Patterns & Conventions

- Entry: `src/index.ts`. Add Vitest tests as `*.test.ts` in `src/**` when extending logic.

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
