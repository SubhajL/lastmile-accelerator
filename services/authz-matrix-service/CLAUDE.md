# Authz Matrix Service - Authorization matrix management

**Technology:** Node.js/TypeScript
**Ports:** REST: 7203, gRPC: 50073
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements authorization matrix service responsibilities per PRD. Manages role-based access control, permission matrices, and authorization policy enforcement across the platform.

## Quick Start

### Development
```bash
cd services/authz-matrix-service
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
authz-matrix-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Authorization logic
│   ├── schemas/          # Zod validation schemas
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/authz.ts` - Core authorization logic
- `src/routes/matrix.ts` - Matrix endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /authz/evaluate` - Evaluate authorization for principal
- `GET /authz/matrix/{roleId}` - Get authorization matrix for role

**gRPC:** See `api/authz.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=authz-matrix-service`
- `SERVICE_PORT=7203`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `JWT_REQUIRED_SCOPES` - Required JWT scopes

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Matrix evaluation:** Cache computed matrices in Redis with 5-min TTL
- **JWT scope checking:** Validate required scopes for each endpoint
- **gRPC integration:** Use for inter-service authz checks

## Related Services

- **jwt-validator:** Validates JWT tokens
- **secrets-env-service:** Stores authorization policies

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-authz-matrix-service.yml`
