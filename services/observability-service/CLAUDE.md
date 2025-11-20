# Observability Service - Centralized Logging, Metrics & Tracing

**Technology:** Go 1.24
**Ports:** REST: 7301, gRPC: 50081
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Observability Service provides centralized logging aggregation, metrics collection, and distributed tracing for the entire LMA platform. It ingests logs from all services, collects and aggregates metrics for dashboards, correlates traces across service boundaries, provides error inbox for production issues, and exposes unified APIs for querying observability data. The service acts as the single source of truth for platform health and performance monitoring.

## Development Commands

### From This Directory
```bash
# Go service commands
go mod download        # Download dependencies
make dev              # Hot-reload with air
make test             # Run tests with coverage
make build            # Build binary
make lint             # Run golangci-lint
make quality          # Run all quality checks (vet, lint, test, build)

# Database migrations
make migration NAME=add_traces_table
make migrate-up
make migrate-down
make migrate-status
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=observability-service
bunx turbo run test --filter=observability-service
bunx turbo run build --filter=observability-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
make quality

# Or individually
go vet ./...
golangci-lint run
go test -race -cover ./...
go build ./...
```

## Architecture

### Directory Structure
```
observability-service/
├── cmd/
│   └── observability-service/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── errorinbox/              # Error inbox for production issues
│   │   ├── service.go
│   │   └── repository.go
│   ├── grpcserver/              # gRPC server implementation
│   ├── handlers/                # HTTP request handlers
│   │   ├── logs.go
│   │   ├── metrics.go
│   │   └── traces.go
│   ├── health/                  # Health check handlers
│   ├── httpjson/                # HTTP JSON utilities
│   ├── integration/             # Integration tests
│   ├── logs/                    # Log ingestion & querying
│   │   ├── ingestor.go
│   │   ├── query.go
│   │   └── storage.go
│   ├── messaging/               # NATS event handling
│   ├── metrics/                 # Metrics collection & aggregation
│   │   ├── collector.go
│   │   ├── aggregator.go
│   │   └── exporter.go
│   ├── middleware/              # HTTP middleware
│   ├── models/                  # Domain models
│   ├── repository/              # Data access layer
│   ├── scheduler/               # Background job scheduler
│   ├── services/                # Business logic
│   ├── storage/                 # Storage abstractions
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── telemetry/               # OpenTelemetry setup
│   └── traces/                  # Trace ingestion & correlation
│       ├── collector.go
│       ├── correlator.go
│       └── query.go
├── migrations/                  # SQL migrations
│   ├── 001_create_logs.sql
│   ├── 002_create_metrics.sql
│   └── 003_create_traces.sql
├── Makefile                     # Build & dev commands
├── go.mod                       # Go dependencies
├── .golangci.yml                # Linter configuration
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Log Ingestor:**
- File: `internal/logs/ingestor.go` - Receives logs from all services
- Pattern: Parses structured JSON logs, extracts trace IDs, stores in PostgreSQL
- Example: Ingests logs via HTTP POST, indexes by timestamp and trace ID

**Metrics Collector:**
- File: `internal/metrics/collector.go` - Collects metrics from Prometheus endpoints
- Pattern: Scrapes `/metrics` from all services, aggregates time-series data
- Example: Collects `http_request_duration_seconds` and computes p50/p95/p99

**Trace Correlator:**
- File: `internal/traces/correlator.go` - Links spans across service boundaries
- Pattern: Builds distributed trace graphs from OpenTelemetry spans
- Example: Correlates HTTP request from projects-service -> db-guardian-service

**Error Inbox:**
- File: `internal/errorinbox/service.go` - Deduplicates and prioritizes production errors
- Pattern: Groups errors by stack trace fingerprint, tracks occurrences
- Example: Groups "Database connection timeout" errors from multiple instances

**Query Engine:**
- Files: `internal/logs/query.go`, `internal/traces/query.go` - Query observability data
- Pattern: Provides flexible search by time range, service, trace ID, log level
- Example: Find all ERROR logs for trace ID abc123 in last 1 hour

### Dependencies

**Core:**
- `jackc/pgx/v5` v5.7.6 - PostgreSQL driver for storing logs, metrics, traces
- `redis/go-redis/v9` v9.16.0 - Redis for caching and real-time aggregation
- `nats-io/nats.go` v1.47.0 - NATS messaging for log/metric streaming
- `golang-jwt/jwt/v5` v5.3.0 - JWT authentication for API access
- `google.golang.org/grpc` v1.75.0 - gRPC server for high-throughput ingestion

**Observability:**
- `go.opentelemetry.io/otel` v1.38.0 - OpenTelemetry SDK (meta-observability)
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` v1.38.0 - OTLP exporter
- `go.opentelemetry.io/otel/metric` v1.38.0 - Metrics SDK
- `go.opentelemetry.io/otel/trace` v1.38.0 - Tracing SDK

**Testing:**
- `DATA-DOG/go-sqlmock` v1.5.2 - SQL mock for unit tests
- Standard library `testing` package

## Code Organization Patterns

### Log Ingestion
✅ **DO:** Use buffered channels for high-throughput log ingestion
```go
// internal/logs/ingestor.go
func (i *Ingestor) Ingest(ctx context.Context, logs []Log) error {
    for _, log := range logs {
        i.logChan <- log  // Non-blocking buffered channel
    }
    return nil
}
```
❌ **DON'T:** Block on synchronous writes; use buffering and batching

### Metrics Aggregation
✅ **DO:** Use time-bucketing for efficient metrics storage
```go
// internal/metrics/aggregator.go
func (a *Aggregator) Aggregate(metric Metric) error {
    bucket := time.Now().Truncate(time.Minute)
    return a.storage.IncrementCounter(bucket, metric)
}
```
❌ **DON'T:** Store individual metric points; aggregate to reduce storage

### Trace Correlation
✅ **DO:** Build trace trees using parent-child span relationships
```go
// internal/traces/correlator.go
func (c *Correlator) CorrelateTrace(traceID string) (*Trace, error) {
    spans := c.repository.GetSpansByTraceID(traceID)
    return buildTraceTree(spans), nil
}
```
❌ **DON'T:** Return flat span lists; always build hierarchical traces

### Error Handling
✅ **DO:** Return structured errors with context
```go
if err != nil {
    return fmt.Errorf("failed to ingest logs for trace %s: %w", traceID, err)
}
```
❌ **DON'T:** Swallow errors from ingestion; log and alert on failures

## API Endpoints

### REST API

**Base URL:** `http://localhost:7301`

**Key Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/logs/ingest` - Ingest batch of logs
- `GET /api/v1/logs/query` - Query logs by filters (time, service, level)
- `POST /api/v1/metrics/ingest` - Ingest metric data points
- `GET /api/v1/metrics/query` - Query aggregated metrics
- `POST /api/v1/traces/ingest` - Ingest OpenTelemetry spans
- `GET /api/v1/traces/{traceId}` - Get full distributed trace
- `GET /api/v1/errorinbox` - Get deduplicated production errors
- `GET /api/v1/errorinbox/{fingerprint}` - Get error details and occurrences

### gRPC API

**Port:** 50081

**Services:**
- `ObservabilityService` - High-throughput log/metric/trace ingestion
  - `IngestLogs` - Batch log ingestion
  - `IngestMetrics` - Metrics streaming
  - `IngestTraces` - OpenTelemetry span ingestion
  - `QueryLogs` - Log search
  - `QueryTraces` - Trace retrieval

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`logs`** - Structured log storage
- Columns: `id`, `timestamp`, `service`, `level`, `message`, `trace_id`, `span_id`, `attributes` (JSONB)
- Indexes: `idx_timestamp`, `idx_trace_id`, `idx_service_level`, `idx_attributes_gin`
- Purpose: Store all application logs with full-text search capability

**`metrics`** - Time-series metric data
- Columns: `id`, `timestamp`, `metric_name`, `value`, `labels` (JSONB), `aggregation_window`
- Indexes: `idx_timestamp_metric`, `idx_labels_gin`
- Purpose: Store aggregated metrics in time buckets

**`traces`** - Distributed trace metadata
- Columns: `trace_id`, `root_span_id`, `service_count`, `duration_ms`, `status`, `started_at`
- Indexes: `idx_trace_id`, `idx_started_at`, `idx_status`
- Purpose: Track high-level trace information

**`spans`** - Individual trace spans
- Columns: `span_id`, `trace_id`, `parent_span_id`, `service`, `operation`, `duration_ms`, `attributes` (JSONB)
- Indexes: `idx_trace_id`, `idx_parent_span_id`, `idx_service`
- Purpose: Store individual operations within traces

**`error_inbox`** - Deduplicated production errors
- Columns: `fingerprint`, `error_type`, `message`, `stack_trace`, `first_seen`, `last_seen`, `occurrence_count`
- Indexes: `idx_fingerprint`, `idx_last_seen`, `idx_occurrence_count`
- Purpose: Track and deduplicate recurring errors

**Migrations:**
- Location: `migrations/`
- Tool: Goose (via Makefile)
- Commands: `make migrate-up`, `make migrate-down`, `make migrate-status`

## Event Handling

**Published Events:**
- `observability.error.detected` - When new error pattern detected
  - Payload: `{fingerprint, error_type, message, service, first_seen}`
- `observability.threshold.exceeded` - When metric exceeds threshold
  - Payload: `{metric_name, threshold, current_value, service}`
- `observability.trace.slow` - When trace duration exceeds SLO
  - Payload: `{trace_id, duration_ms, slo_threshold, services[]}`

**Subscribed Events:**
- `*.*` - All platform events for correlation with logs/traces
- Services publish logs/metrics/traces directly via HTTP/gRPC ingestion

## Testing Strategy

### Unit Tests
- Location: `*_test.go` files colocated with source
- Coverage: Target >80%
- Mock: Database with `go-sqlmock`, Redis, NATS
- Example: `internal/logs/ingestor_test.go`

### Integration Tests
- Location: `internal/integration/`
- Setup: Use Testcontainers for PostgreSQL, Redis
- Pattern: Test full ingestion -> query pipeline

### Running Tests
```bash
# All tests with coverage
make test

# Specific package
go test -v ./internal/logs/...

# With race detector
go test -race ./...

# Integration tests only
go test -v ./internal/integration/...
```

## Configuration

### Environment Variables
```bash
# Service-specific config
LOG_RETENTION_DAYS=30
METRIC_RETENTION_DAYS=90
TRACE_RETENTION_DAYS=7
MAX_BATCH_SIZE=1000
INGESTION_BUFFER_SIZE=10000
ERROR_INBOX_DEDUP_WINDOW=5m

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=observability-service
SERVICE_PORT=7301
GRPC_PORT=50081
ENV=dev
DATABASE_URL=postgresql://user:pass@localhost:5432/observability
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/observability-service/`
- Accessed via: `internal/secrets/vault.go` (if present) or Vault SDK
- Keys: Database credentials, API keys for external observability tools

## Quick Find Commands

### Find Code
```bash
# Find log ingestion logic
rg -n "IngestLogs|ingest.*log" services/observability-service/internal/

# Find metrics collection
rg -n "CollectMetrics|collect.*metric" services/observability-service/internal/

# Find trace correlation
rg -n "CorrelateTrace|correlate" services/observability-service/internal/

# Find error inbox logic
rg -n "ErrorInbox|deduplicat" services/observability-service/internal/

# Find query handlers
rg -n "QueryLogs|QueryMetrics|QueryTraces" services/observability-service/internal/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "observability-service" --glob "docker-compose*.yml" --glob "*.yaml"

# Find services sending observability data
rg -n "localhost:7301|observability.*ingest" services/
```

## Common Gotchas

- **Log Volume:** High-traffic services generate massive log volumes; implement sampling (e.g., 10% of DEBUG logs) and aggressive retention policies
- **Time-Series Bloat:** Metrics with high cardinality labels cause database bloat; limit label combinations and use aggregation
- **Trace Sampling:** Don't store 100% of traces; use head-based or tail-based sampling (e.g., sample errors and slow traces)
- **PostgreSQL Performance:** Large JSONB columns slow queries; use GIN indexes on frequently queried attributes
- **Clock Skew:** Services with unsynchronized clocks produce out-of-order logs; use NTP and handle clock drift
- **Buffering Backpressure:** If ingestion falls behind, buffers fill and logs drop; implement adaptive buffering and alerting
- **Retention Enforcement:** Old data accumulates quickly; run scheduled jobs to delete logs/metrics/traces older than retention period

## Related Services

- **All Services:** Every service sends logs, metrics, and traces to Observability Service
- **notification-service:** Triggers alerts when errors detected or thresholds exceeded
- **projects-service:** Provides project context for filtering observability data
- **db-guardian-service:** Sends migration audit logs for compliance tracking

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-observability-service.yml`
- OpenTelemetry Docs: https://opentelemetry.io/docs/
- Prometheus Best Practices: https://prometheus.io/docs/practices/naming/
