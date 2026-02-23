# Billing Service - Manage billing and usage tracking

**Technology:** Go
**Ports:** REST: 7901, gRPC: 50121
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements billing service responsibilities per PRD. Tracks usage metrics, calculates charges, manages billing cycles, and generates invoices.

## Quick Start

### Development
```bash
cd services/billing-service
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
billing-service/
├── cmd/billing/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Billing logic
│   ├── repository/       # Data access layer
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── migrations/           # SQL migrations
├── Makefile
└── go.mod
```

## Key Files

- `cmd/billing/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/calculator.go` - Charge calculation
- `internal/repository/invoice.go` - Invoice persistence

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /usage/track` - Track usage event
- `GET /invoices/{invoiceId}` - Get invoice details

**gRPC:** See `api/billing.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=billing-service`
- `SERVICE_PORT=7901`
- `GRPC_PORT=50121`
- `DATABASE_URL` - PostgreSQL connection
- `STRIPE_API_KEY` - Stripe integration
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Transaction consistency:** Database transactions for billing operations
- **Event-driven tracking:** NATS events for usage tracking
- **Idempotent calculations:** Replay-safe billing calculations

## Related Services

- **observability-service:** Provides usage metrics
- **notification-service:** Sends billing notifications

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-billing-service.yml`
