# Launch Engine Service - Orchestrate product launch workflows

**Technology:** Node.js/TypeScript
**Ports:** REST: 7502, gRPC: 50102
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements launch engine service responsibilities per PRD. Orchestrates product launch workflows, manages launch timelines, and coordinates multi-service launch operations.

## Quick Start

### Development
```bash
cd services/launch-engine-service
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
launch-engine-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Orchestration logic
│   ├── workflows/        # Launch workflow definitions
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/orchestrator.ts` - Workflow orchestration
- `src/routes/launches.ts` - Launch endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /launches` - Start new launch
- `GET /launches/{launchId}` - Get launch status

**gRPC:** See `api/launch.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=launch-engine-service`
- `SERVICE_PORT=7502`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `NATS_URL` - NATS event streaming

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Workflow state machine:** Tracks launch lifecycle
- **Event-driven coordination:** Uses NATS for inter-service events
- **Rollback capability:** Supports automatic rollback on failure

## Related Services

- **projects-service:** Provides project information
- **notification-service:** Sends launch updates
- **publisher-service:** Publishes launch artifacts

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-launch-engine-service.yml`
