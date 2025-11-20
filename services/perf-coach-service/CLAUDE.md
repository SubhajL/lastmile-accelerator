# Performance Coach Service - Provide performance optimization guidance

**Technology:** Go
**Ports:** REST: 7302, gRPC: 50082
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements performance coach service responsibilities per PRD. Analyzes application performance, identifies bottlenecks, and provides optimization recommendations.

## Quick Start

### Development
```bash
cd services/perf-coach-service
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
perf-coach-service/
├── cmd/perf-coach/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Analysis logic
│   ├── models/           # Domain models
│   ├── analyzer/         # Performance analyzer
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── Makefile
└── go.mod
```

## Key Files

- `cmd/perf-coach/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/analyzer/analyzer.go` - Performance analysis
- `internal/service/coach.go` - Recommendation engine

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /analyze` - Analyze performance data
- `GET /recommendations/{sessionId}` - Get recommendations

**gRPC:** See `api/perfcoach.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=perf-coach-service`
- `SERVICE_PORT=7302`
- `GRPC_PORT=50082`
- `DATABASE_URL` - PostgreSQL connection
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Time series analysis:** Analyzes performance metrics over time
- **Threshold detection:** Identifies performance anomalies
- **Recommendation caching:** Cache common optimization patterns

## Related Services

- **observability-service:** Provides performance metrics
- **notification-service:** Sends performance alerts

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-perf-coach-service.yml`
