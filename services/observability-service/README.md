# Observability Service

Centralized logging, metrics, and distributed tracing service for the LMA platform.

## Overview

The Observability Service provides:
- 📊 **Metrics Collection**: Collects and aggregates metrics from all services
- 📝 **Log Aggregation**: Centralizes logs with structured search capabilities
- 🔍 **Distributed Tracing**: Correlates traces across service boundaries
- 🚨 **SLO Management**: Defines and monitors Service Level Objectives
- ⚠️ **Alert Rules**: Configurable alerting based on metrics and SLOs
- 📮 **Error Inbox**: Deduplicates and prioritizes production errors

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Observability Service                 │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐│
│  │   HTTP   │  │   gRPC   │  │   NATS   │  │  Health ││
│  │  :7301   │  │  :50081  │  │ Consumer │  │   Check ││
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬────┘│
│       │              │              │              │     │
│  ┌────▼─────────────▼──────────────▼──────────────▼───┐│
│  │              Business Logic Layer                   ││
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐ ││
│  │  │   SLO   │ │  Alert  │ │  Error  │ │  Query   │ ││
│  │  │ Service │ │ Service │ │  Inbox  │ │ Service  │ ││
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬─────┘ ││
│  └───────┴───────────┴───────────┴───────────┴────────┘│
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │              Storage Layer                          │ │
│  │  ┌──────────┐  ┌─────────┐  ┌──────────────────┐  │ │
│  │  │PostgreSQL│  │  Redis  │  │ External Sources │  │ │
│  │  │   (Main) │  │ (Cache) │  │ Tempo/Loki/Prom │  │ │
│  │  └──────────┘  └─────────┘  └──────────────────┘  │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- NATS 2.10+

### Local Development

1. **Clone and navigate to service:**
   ```bash
   cd services/observability-service
   ```

2. **Set up environment variables:**
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your local config
   ```

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Run database migrations:**
   ```bash
   make migrate-up
   ```

5. **Start the service:**
   ```bash
   make dev  # Hot-reload development
   # or
   make build && make run  # Production-like
   ```

6. **Verify service is running:**
   ```bash
   curl http://localhost:7301/healthz
   ```

## API Endpoints

### REST API (Port 7301)

#### Health & Metrics
- `GET /healthz` - Health check endpoint
- `GET /metrics` - Prometheus metrics

#### OpenTelemetry Management
- `GET /api/v1/otel-presets` - List available OTel preset configurations
- `POST /api/v1/projects/{projectId}/otel-config` - Configure project OTel settings

#### SLO Management
- `POST /api/v1/projects/{projectId}/slos` - Create SLO
- `GET /api/v1/projects/{projectId}/slos` - List SLOs
- `GET /api/v1/projects/{projectId}/slos/{sloId}` - Get SLO details
- `PUT /api/v1/projects/{projectId}/slos/{sloId}` - Update SLO
- `DELETE /api/v1/projects/{projectId}/slos/{sloId}` - Delete SLO

#### Alert Management
- `POST /api/v1/projects/{projectId}/slos/{sloId}/alerts` - Create alert rule
- `GET /api/v1/projects/{projectId}/alerts` - List project alerts
- `GET /api/v1/projects/{projectId}/alerts/history` - Alert history

#### Error Inbox
- `POST /api/v1/projects/{projectId}/errors/ingest` - Ingest error events
- `GET /api/v1/projects/{projectId}/errors` - List error groups
- `GET /api/v1/projects/{projectId}/errors/{groupId}` - Error group details

#### Observability Queries
- `GET /api/v1/projects/{projectId}/traces/search` - Search traces
- `GET /api/v1/projects/{projectId}/traces/{traceId}` - Get trace details
- `GET /api/v1/projects/{projectId}/logs/query` - Query logs
- `GET /api/v1/projects/{projectId}/golden-signals` - Dashboard metrics

### gRPC API (Port 50081)

See protobuf definitions in `api/observability.proto` for detailed service methods.

## Configuration

### Environment Variables

```bash
# Service Identity
SERVICE_NAME=observability-service
PORT=7301
GRPC_PORT=50081

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/observability
DB_MAX_CONNECTIONS=10

# Cache
REDIS_URL=redis://localhost:6379
REDIS_DB=0

# Messaging
NATS_URL=nats://localhost:4222

# Authentication
OIDC_ISSUER=http://keycloak:8080/realms/lma
JWT_REQUIRED_SCOPES=observability:read,observability:write

# External Observability Sources
TEMPO_URL=http://tempo:3200
LOKI_URL=http://loki:3100
PROMETHEUS_URL=http://prometheus:9090

# Telemetry (Self-monitoring)
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
LOG_LEVEL=info
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v ./internal/slos/...

# Run integration tests
make test-integration
```

## Contributing

1. Follow the coding guidelines in [CLAUDE.md](./CLAUDE.md)
2. Write tests for new features
3. Run `make quality` before committing
4. Use conventional commits

## Monitoring

The service exports metrics at `/metrics` including:
- `observability_slo_evaluations_total` - SLO evaluation count
- `observability_alert_notifications_total` - Alert notifications sent
- `observability_error_events_ingested_total` - Error events processed
- `observability_query_duration_seconds` - Query performance
- Standard HTTP/gRPC metrics

## Troubleshooting

### Service won't start
- Check database connectivity: `psql $DATABASE_URL -c 'SELECT 1'`
- Verify Redis is running: `redis-cli ping`
- Check NATS connection: `nats-cli server check`

### High memory usage
- Adjust `INGESTION_BUFFER_SIZE` (default: 10000)
- Reduce `LOG_RETENTION_DAYS` (default: 30)
- Check for trace sampling configuration

### Slow queries
- Ensure PostgreSQL indexes are created: `make migrate-status`
- Check `METRIC_AGGREGATION_WINDOW` settings
- Review query patterns in slow query log

## Documentation

- [Service Context](./CONTEXT.md) - Service metadata and ownership
- [Development Guide](./CLAUDE.md) - Detailed development patterns
- [API Specification](./AGENTS.md) - Environment and scope matrices