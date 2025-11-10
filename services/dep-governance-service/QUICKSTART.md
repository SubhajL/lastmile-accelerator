# dep-governance-service Quick Start

## Prerequisites
- Rust 1.82+ (toolchain installed)
- PostgreSQL database (required)
- NATS server (optional)
- Redis (optional)

## Quick Setup

### 1. Configure Environment
```bash
# Copy example and edit
cp .env.example .env

# Minimum required configuration
export DATABASE_URL="postgres://user:password@localhost:5432/dep_governance"
```

### 2. Build
```bash
# Development build
cargo build

# Production build (optimized)
cargo build --release
```

### 3. Test
```bash
# Run all tests
cargo test --lib -- --test-threads=1

# Skip database tests if no test DB
cargo test --lib -- --test-threads=1 --skip db_pool
```

### 4. Run
```bash
# Using Makefile
make run

# Direct binary
./target/release/dep_governance_service

# With custom port
SERVICE_PORT=8080 ./target/release/dep_governance_service
```

## Health Check

Once running, verify the service:

```bash
# Liveness
curl http://localhost:7106/healthz
# Response: {"status":"ok"}

# Readiness (requires database)
curl http://localhost:7106/readyz
# Response: {"status":"ready"}

# Metrics
curl http://localhost:7106/metrics
# Response: Prometheus formatted metrics
```

## Development Workflow

### Code Changes
```bash
# Format code
cargo fmt

# Check for errors
cargo check

# Run linter
cargo clippy -- -D warnings

# Run tests
cargo test --lib -- --test-threads=1
```

### Adding New Features
1. Write tests first (TDD)
2. Implement feature
3. Run `cargo test` to verify
4. Run `cargo clippy` for code quality
5. Run `cargo fmt` to format

## Docker

### Build Image
```bash
make image TAG=dev
```

### Run Container
```bash
docker run -d \
  --name dep-governance \
  -p 7106:7106 \
  -e DATABASE_URL="postgres://host.docker.internal:5432/dep_governance" \
  ghcr.io/ORG/REPO-dep-governance-service:dev
```

### Check Logs
```bash
docker logs -f dep-governance
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 7106
lsof -i :7106

# Kill process
kill -9 <PID>

# Or use different port
SERVICE_PORT=7107 ./target/release/dep_governance_service
```

### Database Connection Issues
```bash
# Test connection manually
psql "$DATABASE_URL" -c "SELECT 1"

# Check PostgreSQL is running
pg_isready

# Verify DATABASE_URL format
echo $DATABASE_URL
# Should be: postgres://user:password@host:port/database
```

### Build Issues
```bash
# Clean and rebuild
cargo clean
cargo build --release

# Update dependencies
cargo update
```

## Common Commands

```bash
# Full development cycle
cargo fmt && cargo clippy -- -D warnings && cargo test --lib -- --test-threads=1 && cargo build --release

# Quick test and run
cargo test --lib -- --test-threads=1 && make run

# Build for Docker
make image TAG=$(git rev-parse --short HEAD)

# Generate SBOM
make sbom
```

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `SERVICE_PORT` | No | 7106 | HTTP server port |
| `ENV` | No | dev | Environment (dev/staging/prod) |
| `LOG_LEVEL` | No | info | Logging level |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | - | OpenTelemetry collector |
| `NATS_URL` | No | - | NATS server URL |
| `REDIS_URL` | No | - | Redis connection string |

## Next Steps

- See `PHASE1_COMPLETE.md` for detailed implementation notes
- See `CONTEXT.md` for service purpose and architecture
- See `../../LMA-PRD.md` for overall system design

---

**Need Help?** Check the test files in `src/*/tests/` for usage examples.
