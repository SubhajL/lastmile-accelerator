# Rate Limit Service - Enforce API rate limiting policies

**Technology:** Go
**Ports:** REST: 7204, gRPC: 50074
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements rate limit service responsibilities per PRD. Enforces rate limiting policies, manages quota allocations, and provides rate limit information to other services.

## Quick Start

### Development
```bash
cd services/rate-limit-service
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
rate-limit-service/
├── cmd/rate-limit/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # Rate limiting logic
│   ├── models/           # Domain models
│   ├── cache/            # Redis caching
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── Makefile
└── go.mod
```

## Key Files

- `cmd/rate-limit/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/service/limiter.go` - Rate limiting logic
- `internal/cache/redis.go` - Redis integration

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /check` - Check rate limit status
- `GET /quota/{clientId}` - Get quota info

**gRPC:** See `api/ratelimit.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=rate-limit-service`
- `SERVICE_PORT=7204`
- `GRPC_PORT=50074`
- `REDIS_URL` - Redis connection
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Token bucket algorithm:** Implements rate limiting via token bucket
- **Redis persistence:** Quota state stored in Redis
- **Policy caching:** Load limit policies with periodic refresh

## Related Services

- **authz-matrix-service:** Defines rate limit tiers per role
- **observability-service:** Reports rate limit metrics

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-rate-limit-service.yml`
