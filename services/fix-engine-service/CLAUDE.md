# Fix Engine Service - Generate and apply automated fixes

**Technology:** Node.js/TypeScript
**Ports:** REST: 7151, gRPC: 50067
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements fix engine service responsibilities per PRD. Generates automated fixes for code issues, manages fix proposals, and applies safe transformations to codebases.

## Quick Start

### Development
```bash
cd services/fix-engine-service
pnpm install
pnpm dev
```

### Testing
```bash
pnpm test                 # Run all tests
pnpm test:watch          # Watch mode
pnpm test:coverage       # Generate coverage
```

### Pre-PR
```bash
pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

## Directory Structure

```
fix-engine-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Fix generation logic
│   ├── transforms/       # Code transformations
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/fixGenerator.ts` - Fix generation
- `src/routes/fixes.ts` - Fix endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /fixes/generate` - Generate fix for issue
- `GET /fixes/{fixId}` - Get fix details

**gRPC:** See `api/fix.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=fix-engine-service`
- `SERVICE_PORT=7151`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `FIX_STRATEGIES_URL` - URL to fix strategies

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **AST manipulation:** Transform code via abstract syntax trees
- **Diff generation:** Create readable diffs of proposed fixes
- **Safety validation:** Ensure fixes don't break compilation

## Related Services

- **ai-debugger-service:** Detects issues for fixing
- **test-lab-service:** Validates fixes with testing

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-fix-engine-service.yml`
