# Legal Automator Service - Automate legal document generation and compliance

**Technology:** Node.js/TypeScript
**Ports:** REST: 7503, gRPC: 50103
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements legal automator service responsibilities per PRD. Automates legal document generation, manages compliance templates, and enforces legal requirements across products.

## Quick Start

### Development
```bash
cd services/legal-automator-service
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
legal-automator-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Legal logic
│   ├── templates/        # Legal document templates
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/documentGenerator.ts` - Document generation
- `src/routes/documents.ts` - Document endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /documents/generate` - Generate legal document
- `GET /documents/{docId}` - Get generated document

**gRPC:** See `api/legal.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=legal-automator-service`
- `SERVICE_PORT=7503`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `LEGAL_TEMPLATES_URL` - URL to legal templates

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Template management:** Load and cache legal templates
- **Data substitution:** Safe variable substitution in templates
- **Audit trail:** Track all document generation events

## Related Services

- **projects-service:** Provides project context
- **notification-service:** Notifies of document generation

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-legal-automator-service.yml`
