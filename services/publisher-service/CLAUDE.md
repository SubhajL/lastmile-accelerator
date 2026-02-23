# Publisher Service - Publish artifacts and releases

**Technology:** Go
**Ports:** REST: 7201, gRPC: 50071
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements publisher service responsibilities per PRD. Manages artifact publishing, release coordination, and distribution across deployment environments.

## Quick Start

### Development
```bash
cd services/publisher-service
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
publisher-service/
├── cmd/publisher/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Business logic
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/publisher/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/publisher.go` - Publishing logic
- `api/publisher.proto` - gRPC definitions

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /publish` - Publish artifact
- `GET /releases/{releaseId}` - Get release info

**gRPC:** See `api/publisher.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=publisher-service`
- `SERVICE_PORT=7201`
- `GRPC_PORT=50071`
- `DATABASE_URL` - PostgreSQL connection
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Transaction management:** Multi-step publish operations use transactions
- **Event publishing:** NATS events for publish milestones
- **Artifact storage:** S3 integration for artifact persistence

## Related Services

- **launch-engine-service:** Orchestrates product launches
- **notification-service:** Notifies stakeholders

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-publisher-service.yml`
