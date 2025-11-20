# MVP Validator Service - Validate minimum viable product specifications

**Technology:** Node.js/TypeScript
**Ports:** REST: 7402, gRPC: 50092
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements MVP validation service responsibilities per PRD. Validates product specifications against minimum viability requirements, feature completeness, and quality gates.

## Quick Start

### Development
```bash
cd services/mvp-validator-service
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
mvp-validator-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Validation logic
│   ├── schemas/          # Zod validation schemas
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/validator.ts` - Core validation logic
- `src/routes/validate.ts` - Validation endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /validate/spec` - Validate product specification
- `GET /validate/requirements` - Get MVP requirements

**gRPC:** See `api/mvp.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=mvp-validator-service`
- `SERVICE_PORT=7402`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `VALIDATION_RULES_URL` - URL to validation rules

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Validation rules:** Cached in memory with periodic refresh
- **Async validation:** Long-running validations use job queue
- **Result reporting:** Detailed validation reports with severity levels

## Related Services

- **projects-service:** Provides product specifications
- **notification-service:** Sends validation result notifications

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-mvp-validator-service.yml`
