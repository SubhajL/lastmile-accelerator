> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/test-lab-service-spec.md
- Progress:    ../../documentation/features/active/test-lab-service-progress.md
- Planned:     ../../documentation/features/planned/test-lab-service-spec.md

## Package Identity

- Node/TypeScript service orchestrating test scaffolds, previews, and runs.

## Setup & Run

```bash
# From repo root
bun install
bunx turbo run build --filter=test-lab-service
bunx turbo run typecheck --filter=test-lab-service
bunx turbo run test --filter=test-lab-service
make run
```

## Patterns & Conventions

- REST routes in `src/routes/*.ts` (e.g., `src/routes/scaffolds.routes.ts`).
- DB layer in `src/repo/**`; schemas in `src/schemas/**`.
- Events/subscribers: `src/events/subscribers.ts`.
- Tests split into `src/__tests__/unit/**` and `src/__tests__/integration/**`.

## Touch Points / Key Files

- `src/app.ts`, `src/index.ts`
- `src/repo/scaffolds.pg.repo.ts`, `src/schemas/scaffolds.ts`
- `src/events/subscribers.ts`
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'describe\(|it\(' src/__tests__
grep -R -nE 'fastify\.(get|post|put|delete)\(' src/routes
```

## Pre-PR Checks

```bash
bunx turbo run typecheck test build --filter=test-lab-service
```
