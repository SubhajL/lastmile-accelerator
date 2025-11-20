# Positioning Engine Service - Product positioning and market analysis

**Technology:** Python/FastAPI
**Ports:** REST: 7501, gRPC: 50101
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements positioning engine service responsibilities per PRD. Analyzes market positioning, manages product positioning statements, and provides competitive intelligence.

## Quick Start

### Development
```bash
cd services/positioning-engine-service
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
positioning-engine-service/
├── src/
│   ├── main.py            # Entry point
│   ├── config.py          # Configuration
│   ├── routes/            # FastAPI routes
│   ├── services/          # Positioning logic
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
- `src/routes/positioning.py` - Positioning endpoints
- `src/services/analyzer.py` - Market analysis
- `src/models/position.py` - Positioning data model

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /positioning/analyze` - Analyze positioning
- `GET /positioning/{positionId}` - Get positioning details

**gRPC:** See `api/positioning.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=positioning-engine-service`
- `SERVICE_PORT=7501`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `DATABASE_URL` - PostgreSQL connection
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `MARKET_DATA_API_KEY` - Market data API key

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `tests/test_*.py`
- **Run:** `pytest tests/`
- **Coverage:** `pytest tests/ --cov=src`

## Common Patterns

- **Market analysis:** Analyze competitive landscape
- **Sentiment analysis:** Extract positioning sentiment from data
- **Trend detection:** Identify market trends using time series
- **Caching:** Cache market analysis in Redis with daily TTL

## Related Services

- **projects-service:** Provides product information
- **notification-service:** Sends positioning updates

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-positioning-engine-service.yml`
