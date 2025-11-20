# Agent Ingest Service - Ingest and process agent data streams

**Technology:** Go
**Ports:** REST: 7053, gRPC: 50043
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements agent ingest service responsibilities per PRD. Receives data from distributed agents, validates schemas, and routes to appropriate processing services.

## Quick Start

### Development
```bash
cd services/agent-ingest-service
make dev
```

### Testing
```bash
make test         # Run tests with coverage
go test -v ./... # Verbose output
```

### Pre-PR
```bash
make quality  # Runs vet, lint, test, build
```

## Directory Structure

```
agent-ingest-service/
├── cmd/agent-ingest/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Ingest logic
│   ├── validators/       # Schema validators
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/agent-ingest/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/ingest.go` - Ingest logic
- `internal/validators/validator.go` - Schema validation

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /ingest` - Receive agent data
- `GET /status/{agentId}` - Get agent status

**gRPC:** See `api/agentingest.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=agent-ingest-service`
- `SERVICE_PORT=7053`
- `GRPC_PORT=50043`
- `DATABASE_URL` - PostgreSQL connection
- `NATS_URL` - NATS event streaming
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Schema validation:** Validate incoming agent data against schemas
- **Event routing:** Route validated data via NATS topics
- **Deduplication:** Prevent duplicate data ingestion

## Related Services

- **snapshot-orchestrator-service:** Orchestrates processing
- **observability-service:** Tracks ingest metrics

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-agent-ingest-service.yml`
