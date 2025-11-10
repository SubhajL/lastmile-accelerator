> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/notification-service-spec.md
- Progress:    ../../documentation/features/active/notification-service-progress.md
- Planned:     ../../documentation/features/planned/notification-service-spec.md

## Package Identity

- Node/TypeScript Fastify microservice for notifications (REST port 7902; see service_catalog.yaml).

## Setup & Run

```bash
# from repo root
pnpm i
pnpm --filter notification-service build
pnpm --filter notification-service typecheck
pnpm --filter notification-service test
make run   # or: pnpm --filter notification-service start
```

## Patterns & Conventions

- Entry: `src/index.ts` bootstraps `src/app.ts`.
- Colocated tests under `src/**` (e.g., `src/app.test.ts`, `src/templates/validate.test.ts`).
- Feature areas:
  - Channels: `src/channels/**`
  - Notifications: `src/notifications/**`
  - Templates: `src/templates/**`
  - Events: `src/events/**`
  - Metrics/OTel: `src/metrics/**`, `src/telemetry.ts`

## Touch Points / Key Files

- Server: `src/app.ts`, `src/index.ts`
- Routing/worker: `src/worker.ts`, `src/routing/**`
- Events: `src/events/nats.ts`
- Config: `src/config.ts`, `vitest.config.ts`, `tsconfig.json`, `package.json`
- Deploy: `helm/templates/deployment.yaml`

## JIT Index Hints

```bash
grep -R -nE 'fastify\.(get|post|put|delete)\(' src
grep -R -nE 'describe\(|it\(' src
grep -R -n 'export ' src | head
```

## Common Gotchas

- Integration tests expect Redis/Postgres/NATS; set env vars or run unit tests only.
- Keep `tsc --noEmit` green (`pnpm --filter notification-service typecheck`).

## Pre-PR Checks

```bash
pnpm --filter notification-service typecheck && pnpm --filter notification-service test && pnpm --filter notification-service build
```
