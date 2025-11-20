# AI Debugger Service - AI-powered debugging and issue detection

**Technology:** Python/FastAPI
**Ports:** REST: 7102, gRPC: 50062
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements AI debugger service responsibilities per PRD. Uses AI to analyze code, detect issues, identify root causes, and suggest fixes for software problems.

## Quick Start

### Development
```bash
cd services/ai-debugger-service
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
ai-debugger-service/
├── src/
│   ├── main.py            # Entry point
│   ├── config.py          # Configuration
│   ├── routes/            # FastAPI routes
│   ├── services/          # Debugging logic
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
- `src/routes/debug.py` - Debug endpoints
- `src/services/analyzer.py` - AI analysis logic
- `src/models/issue.py` - Issue data model

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /debug/analyze` - Analyze code for issues
- `GET /debug/{issueId}` - Get issue details

**gRPC:** See `api/debugger.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=ai-debugger-service`
- `SERVICE_PORT=7102`
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `DATABASE_URL` - PostgreSQL connection
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets
- `AI_MODEL_ENDPOINT` - AI model service endpoint

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `tests/test_*.py`
- **Run:** `pytest tests/`
- **Coverage:** `pytest tests/ --cov=src`

## Common Patterns

- **Async operations:** FastAPI async/await for I/O
- **Type hints:** All functions use type hints
- **Error handling:** Custom exception classes for domain errors
- **AI integration:** Call external AI model service

## Related Services

- **fix-engine-service:** Generates fixes for detected issues
- **test-lab-service:** Validates detected issues with tests

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-ai-debugger-service.yml`
