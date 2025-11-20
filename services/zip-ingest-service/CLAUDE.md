# ZIP Ingest Service - Process uploaded ZIP files and extract content

**Technology:** Node.js/TypeScript
**Ports:** REST: 7052, gRPC: 50042
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements ZIP ingest service responsibilities per PRD. Processes uploaded ZIP files, extracts content, validates structure, and forwards to appropriate processing services.

## Quick Start

### Development
```bash
cd services/zip-ingest-service
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
zip-ingest-service/
├── src/
│   ├── __tests__/        # Test suites (Vitest)
│   ├── index.ts          # Entry point
│   ├── app.ts            # Fastify app setup
│   ├── routes/           # REST endpoint handlers
│   ├── services/         # ZIP processing logic
│   ├── validators/       # Content validators
│   └── types/            # TypeScript types
├── package.json
└── Dockerfile
```

## Key Files

- `src/index.ts` - Entry point
- `src/app.ts` - Fastify app setup
- `src/services/zipProcessor.ts` - ZIP processing
- `src/routes/ingest.ts` - Ingest endpoints

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /ingest/upload` - Upload and process ZIP
- `GET /ingest/{ingestId}` - Get ingest status

**gRPC:** See `api/ingest.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=zip-ingest-service`
- `SERVICE_PORT=7052`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `MAX_UPLOAD_SIZE` - Maximum ZIP file size
- `TEMP_STORAGE_PATH` - Temporary extraction directory

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `src/__tests__/*.test.ts`
- **Run:** `pnpm test`
- **Coverage:** `pnpm test:coverage`

## Common Patterns

- **Stream processing:** Handle large ZIP files via streams
- **Content validation:** Scan for malware and validate structure
- **Cleanup:** Temporary file management and TTL expiry

## Related Services

- **agent-ingest-service:** Processes extracted agent data
- **snapshot-orchestrator-service:** Orchestrates processing pipeline

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-zip-ingest-service.yml`
