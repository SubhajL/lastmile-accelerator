# Scaffold Secure Service - Generate secure project scaffolding

**Technology:** Node.js/TypeScript
**Ports:** REST: 7103
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements scaffold secure service responsibilities per PRD. Generates secure project scaffolds with built-in security best practices and compliance requirements.

## Quick Start

### Development
```bash
cd services/scaffold-secure-service
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
scaffold-secure-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Scaffolding logic
│   ├── templates/        # Secure templates
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/scaffolder.ts` - Scaffold generation
- `src/routes/scaffold.ts` - Scaffold endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /scaffold/generate` - Generate scaffold
- `GET /scaffold/{scaffoldId}` - Get generated scaffold

## Configuration

Key environment variables:
- `SERVICE_NAME=scaffold-secure-service`
- `SERVICE_PORT=7103`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `TEMPLATE_PATH` - Path to secure templates

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Template injection:** Secure variable substitution
- **Compliance checking:** Validate scaffold against security standards
- **Asset generation:** File templating and artifact creation

## Related Services

- **secrets-env-service:** Provisions secure secrets
- **projects-service:** Stores generated scaffolds

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-scaffold-secure-service.yml`
