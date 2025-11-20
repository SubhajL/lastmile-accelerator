# Snapshot Orchestrator Service - Orchestrate snapshot processing workflows

**Technology:** Go
**Ports:** REST: 7054, gRPC: 50044
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements snapshot orchestrator service responsibilities per PRD. Orchestrates multi-step snapshot processing workflows, coordinates parallel tasks, and manages snapshot lifecycle.

## Quick Start

### Development
```bash
cd services/snapshot-orchestrator-service
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
snapshot-orchestrator-service/
├── cmd/snapshot/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Orchestration logic
│   ├── workflows/        # Workflow definitions
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/snapshot/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/orchestrator.go` - Workflow orchestration
- `internal/workflows/processor.go` - Workflow execution

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /snapshots/process` - Start snapshot processing
- `GET /snapshots/{snapshotId}` - Get snapshot status

**gRPC:** See `api/snapshot.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=snapshot-orchestrator-service`
- `SERVICE_PORT=7054`
- `GRPC_PORT=50044`
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

- **State machine workflow:** Manages workflow state transitions
- **Parallel task execution:** Coordinates concurrent processing tasks
- **Failure handling:** Implements rollback and retry logic

## Related Services

- **agent-ingest-service:** Receives data to process
- **observability-service:** Tracks workflow metrics

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-snapshot-orchestrator-service.yml`
