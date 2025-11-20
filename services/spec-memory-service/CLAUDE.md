# Spec Memory Service - Store and retrieve specification memory

**Technology:** Go
**Ports:** REST: 7101, gRPC: 50061
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements spec memory service responsibilities per PRD. Manages specification versioning, stores architectural decisions, and provides context-aware spec lookups for AI services.

## Quick Start

### Development
```bash
cd services/spec-memory-service
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
spec-memory-service/
├── cmd/spec-memory/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Memory logic
│   ├── repository/       # Data access layer
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/spec-memory/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/memory.go` - Memory management
- `internal/repository/spec.go` - Specification storage

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /specs/store` - Store specification
- `GET /specs/{specId}` - Retrieve specification

**gRPC:** See `api/specmemory.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=spec-memory-service`
- `SERVICE_PORT=7101`
- `GRPC_PORT=50061`
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis for caching
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Version control:** Track spec revisions and history
- **Semantic search:** Find specs by description/keywords
- **Context injection:** Provide relevant specs to AI services
- **TTL-based caching:** Cache frequently accessed specs in Redis

## Related Services

- **ai-debugger-service:** Consumes spec context
- **projects-service:** Stores spec relationships

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-spec-memory-service.yml`
