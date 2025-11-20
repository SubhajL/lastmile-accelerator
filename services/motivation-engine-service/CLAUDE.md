# Motivation Engine Service - Manage developer motivation and engagement

**Technology:** Node.js/TypeScript
**Ports:** REST: 7601, gRPC: 50111
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements motivation engine service responsibilities per PRD. Tracks developer engagement, manages incentive systems, and provides motivation-based recommendations.

## Quick Start

### Development
```bash
cd services/motivation-engine-service
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
motivation-engine-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # Engagement logic
│   ├── models/           # Data models
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/engagement.ts` - Engagement tracking
- `src/routes/motivation.ts` - Motivation endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /motivation/track` - Track engagement event
- `GET /motivation/status/{userId}` - Get user motivation status

**gRPC:** See `api/motivation.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=motivation-engine-service`
- `SERVICE_PORT=7601`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `REDIS_URL` - Redis for engagement tracking

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Engagement scoring:** Real-time calculation of motivation metrics
- **Event tracking:** NATS pub/sub for engagement events
- **Recommendation engine:** AI-based recommendations caching

## Related Services

- **notification-service:** Sends motivation notifications
- **projects-service:** Tracks project contributions

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-motivation-engine-service.yml`
