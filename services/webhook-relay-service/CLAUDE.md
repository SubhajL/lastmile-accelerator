# Webhook Relay Service - Relay and manage webhooks

**Technology:** Go
**Ports:** REST: 7903, gRPC: 50123
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements webhook relay service responsibilities per PRD. Manages webhook registrations, queues events, and reliably delivers webhooks to external endpoints with retry logic.

## Quick Start

### Development
```bash
cd services/webhook-relay-service
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
webhook-relay-service/
├── cmd/webhook/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Webhook logic
│   ├── queue/            # Event queueing
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/webhook/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/relay.go` - Webhook relay logic
- `internal/queue/queue.go` - Event queueing

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /webhooks/register` - Register webhook
- `GET /webhooks/{hookId}` - Get webhook details

**gRPC:** See `api/webhook.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=webhook-relay-service`
- `SERVICE_PORT=7903`
- `GRPC_PORT=50123`
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

- **Exponential backoff:** Retry failed deliveries with backoff
- **Event deduplication:** Prevent duplicate webhook deliveries
- **Delivery tracking:** Store delivery history and results

## Related Services

- **notification-service:** Sends delivery notifications
- **observability-service:** Tracks webhook metrics

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-webhook-relay-service.yml`
