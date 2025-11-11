> Global Rules (must‑read)
> - Follow root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Package Identity

- Next.js 14 app (App Router) for ZIP uploader UX.
- Toolchain: pnpm for install/build; Node 20; TypeScript strict.
- Note: The monorepo prefers Bun for Node services, but frontends use pnpm consistently.

## Setup & Run

```bash
pnpm i
pnpm dev           # local dev server
pnpm build         # production build
pnpm start         # run built app
pnpm run typecheck # tsc -p .
```

## Pre-PR Checks

```bash
pnpm run typecheck && pnpm build
```

## CI Notes

- Dockerfile uses pnpm (`corepack enable && pnpm i --frozen-lockfile`).
- GitHub Actions sets up Node 20 and pnpm; builds the app.
- Optional: Playwright smoke tests may run in CI when configured.

## JIT Index Hints

```bash
rg -n "next|app/|pages/|components/" src || true
```

