> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/projects-service-spec.md
- Progress:    ../../documentation/features/active/projects-service-progress.md
- Planned:     ../../documentation/features/planned/projects-service-spec.md

## Package Identity

- Node/TypeScript Fastify microservice managing projects, tenants, environments, and members.

## Setup & Run

```bash
pnpm i
pnpm --filter projects-service build
pnpm --filter projects-service typecheck || true
pnpm --filter projects-service test
make run
```

## Patterns & Conventions

- Routes live in `src/routes/*.ts` (e.g., `src/routes/projects.ts`, `src/routes/tenants.ts`, `src/routes/environments.ts`, `src/routes/members.ts`).
- Tests are split:
  - Unit: `src/__tests__/unit/**`
  - Integration: `src/__tests__/integration/**`
- Observability: `src/otel.ts`; NATS: `src/nats.ts`; DB access in `src/db.ts` and `src/db/**`.

## Touch Points / Key Files

- Entry: `src/index.ts`
- Config/Obs: `src/otel.ts`, `src/logger.ts`
- Routes: `src/routes/*.ts`
- Tests: `src/__tests__/unit/**`, `src/__tests__/integration/**`
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'fastify\.(get|post|put|delete)\(' src/routes
grep -R -n '__tests__/(integration|unit)' src
```

## Common Gotchas

- Integration tests require Postgres and NATS; ensure env and migrations are applied (`pnpm --filter projects-service migrate`).

## Pre-PR Checks

```bash
pnpm --filter projects-service test && pnpm --filter projects-service build
```
