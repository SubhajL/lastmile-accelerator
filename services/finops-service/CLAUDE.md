# FinOps Service - Financial operations and cost optimization

**Technology:** Python/FastAPI
**Ports:** REST: 7401, gRPC: 50091
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements FinOps service responsibilities per PRD. Tracks cloud spending, analyzes cost trends, provides cost optimization recommendations, and manages budgets.

## Quick Start

### Development
```bash
cd services/finops-service
python -m venv venv
source venv/bin/activate  # or `venv\Scripts\activate` on Windows
pip install -r requirements.txt
python src/main.py
```

### Testing
```bash
pytest tests/                # Run all tests
pytest tests/ -v            # Verbose output
pytest tests/ --cov         # Generate coverage
```

### Pre-PR
```bash
python -m mypy src && pytest tests/ --cov && python -m black . --check
```

## Directory Structure

```
finops-service/
├── src/
│   ├── main.py            # Entry point
│   ├── config.py          # Configuration
│   ├── routes/            # FastAPI routes
│   ├── services/          # FinOps logic
│   ├── models/            # Data models
│   ├── db/                # Database layer
│   └── __init__.py
├── tests/                 # pytest tests
├── requirements.txt       # Dependencies
├── pyproject.toml         # Python project config
└── Dockerfile
```

## Key Files

- `src/main.py` - Entry point
- `src/routes/costs.py` - Cost tracking endpoints
- `src/services/optimizer.py` - Cost optimization analysis
- `src/models/spending.py` - Spending data model

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `GET /costs/summary` - Get cost summary
- `POST /costs/track` - Track cost event

**gRPC:** See `api/finops.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=finops-service`
- `SERVICE_PORT=7401`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `DATABASE_URL` - PostgreSQL connection
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `CLOUD_PROVIDER` - AWS/GCP/Azure provider
- `CLOUD_API_KEY` - Cloud provider API key

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `tests/test_*.py`
- **Run:** `pytest tests/`
- **Coverage:** `pytest tests/ --cov=src`

## Common Patterns

- **Time series analysis:** Analyze spending patterns over time
- **Budget alerts:** Trigger alerts when budget threshold exceeded
- **Multi-cloud support:** Aggregate costs from multiple cloud providers
- **Caching:** Cache cost summaries in Redis with hourly TTL

## Related Services

- **billing-service:** Provides usage data for cost calculation
- **notification-service:** Sends cost alerts

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-finops-service.yml`
