# Dev - Local Development Environment

**Parent Context:** This extends [../CLAUDE.md](../CLAUDE.md)

This directory contains scripts and configuration for running the complete LMA platform locally using Docker Compose and hot-reload development servers.

## Quick Start

```bash
# From this directory (dev/)
./dev.sh start      # Start infrastructure + all Sprint 1 services
./dev.sh status     # Check what's running
./dev.sh logs       # View infrastructure logs
./dev.sh stop       # Stop everything
```

## Directory Structure

```
dev/
├── dev.sh                        # Main orchestration script
├── Procfile                      # Process definitions for Hivemind
├── README.md                     # Detailed setup guide
├── .env.local                    # Local environment variables
├── docker-compose.infra.yml      # Infrastructure services (if separate)
└── scripts/                      # Helper scripts
```

## Infrastructure Stack

**Location:** `../lma-devstack-compose-gitea4001/docker-compose.yml`

### Services

**Databases:**
- **PostgreSQL** - Port 55432
  - Database: `lma_dev`
  - User: `lma`
  - Password: `lma123`

**Caching:**
- **Redis** - Port 4050
  - No authentication in dev
  - DB 0 for most services

**Messaging:**
- **NATS** - Port 4222
  - Cluster ID: `lma-cluster`
  - No authentication in dev

**Secrets:**
- **Vault** - Port 8200
  - Root token: `dev-root-token`
  - Dev mode (in-memory storage)

**Storage:**
- **MinIO (S3-compatible)** - Port 9000
  - Console: Port 9001
  - Access Key: `minioadmin`
  - Secret Key: `minioadmin`

**Authentication:**
- **Keycloak (OIDC)** - Port 8080
  - Admin: `admin` / `admin`
  - Realm: `lma`

**Observability:**
- **Prometheus** - Port 9090
  - Scrapes metrics from all services
- **Grafana** - Port 3000
  - User: `admin` / `admin`
- **OpenTelemetry Collector** - Port 4318 (HTTP), 4317 (gRPC)
  - Receives traces from services

**Development Tools:**
- **Gitea (Git server)** - Port 4001
  - Admin: `gitea` / `gitea123`
- **MailHog (Email testing)** - Port 8025
  - SMTP: Port 1025
  - Web UI: Port 8025

## Development Workflow

### 1. First-Time Setup

```bash
# Install dependencies
./dev.sh install

# This installs:
# - Hivemind (tmux-based process manager)
# - air (Go hot-reload)
# - cargo-watch (Rust hot-reload)
# - tsx (Node hot-reload)
```

### 2. Start Infrastructure

```bash
cd dev
./dev.sh start

# Or manually:
cd ../lma-devstack-compose-gitea4001
docker-compose up -d
```

**Verify Infrastructure:**
```bash
./dev.sh deps

# Or manually check:
curl http://localhost:55432  # PostgreSQL
redis-cli -p 4050 ping       # Redis
curl http://localhost:8200   # Vault
curl http://localhost:9000   # MinIO
```

### 3. Start Services

**Using Hivemind (recommended):**
```bash
cd dev
hivemind Procfile
```

**Procfile defines 7 Sprint 1 services:**
```yaml
projects:       cd ../services/projects-service && pnpm dev
observability:  cd ../services/observability-service && make dev
notification:   cd ../services/notification-service && pnpm dev
dep-governance: cd ../services/dep-governance-service && cargo watch -x run
db-guardian:    cd ../services/db-guardian-service && make dev
test-lab:       cd ../services/test-lab-service && pnpm dev
secrets-env:    cd ../services/secrets-env-service && make dev
```

**Start Individual Service:**
```bash
# Node service
cd ../services/projects-service
pnpm install
pnpm dev

# Go service
cd ../services/db-guardian-service
make dev

# Rust service
cd ../services/dep-governance-service
cargo watch -x run
```

### 4. Access Services

**API Endpoints:**
```
http://localhost:7002  # projects-service
http://localhost:7301  # observability-service
http://localhost:7902  # notification-service
http://localhost:7106  # dep-governance-service
http://localhost:7105  # db-guardian-service
http://localhost:7202  # test-lab-service
http://localhost:7104  # secrets-env-service
```

**Health Checks:**
```bash
curl http://localhost:7002/healthz
curl http://localhost:7105/healthz
# ... etc for each service
```

**Metrics:**
```bash
curl http://localhost:7002/metrics
curl http://localhost:7105/metrics
# ... etc for each service
```

### 5. Development Cycle

**Make Changes:**
- Edit code in service directory
- Hot-reload kicks in automatically:
  - **Node:** sub-second reload with tsx watch
  - **Go:** ~2s rebuild with air
  - **Rust:** ~3-5s rebuild with cargo-watch

**Run Tests:**
```bash
# Node
cd services/projects-service
pnpm test

# Go
cd services/db-guardian-service
make test

# Rust
cd services/dep-governance-service
cargo test --all
```

**Check Logs:**
```bash
# Hivemind - all services in tmux panes
# Use Ctrl+B then arrow keys to navigate panes

# Individual service logs
cd services/projects-service
pnpm dev 2>&1 | tee dev.log
```

### 6. Database Operations

**Connect to PostgreSQL:**
```bash
psql postgresql://lma:lma123@localhost:55432/lma_dev
```

**Run Migrations:**
```bash
# Node service
cd services/projects-service
pnpm db:migrate:up

# Go service
cd services/db-guardian-service
make migrate-up

# Rust service
cd services/dep-governance-service
sqlx migrate run
```

**View Database:**
```bash
# List tables
psql postgresql://lma:lma123@localhost:55432/lma_dev -c "\dt"

# Query data
psql postgresql://lma:lma123@localhost:55432/lma_dev -c "SELECT * FROM projects;"
```

### 7. Testing Event Flow

**Publish Event (using NATS CLI):**
```bash
nats pub -s localhost:4222 project.created '{"projectId":"123","userId":"user1"}'
```

**Subscribe to Events:**
```bash
nats sub -s localhost:4222 "project.*"
```

**Check Event Handlers:**
- Services log received events
- Check service logs for event processing

### 8. Testing Email

**Send Email:**
```bash
# From notification-service
curl -X POST http://localhost:7902/notifications \
  -H "Content-Type: application/json" \
  -d '{"type":"email","to":"test@example.com","subject":"Test","body":"Hello"}'
```

**View Email in MailHog:**
```bash
open http://localhost:8025
```

### 9. Testing S3 (MinIO)

**Upload File:**
```bash
# Using AWS CLI with MinIO
aws --endpoint-url http://localhost:9000 \
  s3 cp test.txt s3://lma-dev/test.txt
```

**View in MinIO Console:**
```bash
open http://localhost:9001
```

### 10. Observability

**View Metrics in Prometheus:**
```bash
open http://localhost:9090
# Query: http_requests_total{service="projects-service"}
```

**View Dashboards in Grafana:**
```bash
open http://localhost:3000
```

**View Traces:**
- OpenTelemetry Collector receives traces on port 4318
- Export to Jaeger or Tempo backend for visualization

## Environment Variables

**File:** `.env.local` (gitignored)

```bash
# Database
DATABASE_URL=postgresql://lma:lma123@localhost:55432/lma_dev

# Redis
REDIS_URL=redis://localhost:4050

# NATS
NATS_URL=nats://localhost:4222

# Vault
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=dev-root-token

# MinIO (S3)
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET=lma-dev

# Keycloak (OIDC)
OIDC_ISSUER=http://localhost:8080/realms/lma
OIDC_CLIENT_ID=lma-dev
OIDC_CLIENT_SECRET=dev-secret

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=debug

# Service Ports (reference)
PROJECTS_SERVICE_URL=http://localhost:7002
OBSERVABILITY_SERVICE_URL=http://localhost:7301
NOTIFICATION_SERVICE_URL=http://localhost:7902
DB_GUARDIAN_SERVICE_URL=http://localhost:7105
DEP_GOVERNANCE_SERVICE_URL=http://localhost:7106
TEST_LAB_SERVICE_URL=http://localhost:7202
SECRETS_ENV_SERVICE_URL=http://localhost:7104
```

## dev.sh Script Reference

### Commands

```bash
./dev.sh start      # Start infra + all services (Hivemind)
./dev.sh stop       # Stop all services + infra
./dev.sh restart    # Restart services (keep infra running)
./dev.sh status     # Show running processes
./dev.sh logs       # Tail infrastructure logs
./dev.sh deps       # Check if dependencies are ready
./dev.sh clean      # Stop and delete all volumes
./dev.sh install    # Install dev tools (hivemind, air, cargo-watch)
```

### Script Logic

```bash
#!/bin/bash

start() {
  echo "Starting infrastructure..."
  cd ../lma-devstack-compose-gitea4001
  docker-compose up -d

  echo "Waiting for services to be ready..."
  wait_for_postgres
  wait_for_redis
  wait_for_vault

  echo "Starting application services..."
  cd ../dev
  hivemind Procfile
}

stop() {
  echo "Stopping application services..."
  pkill -f hivemind

  echo "Stopping infrastructure..."
  cd ../lma-devstack-compose-gitea4001
  docker-compose down
}

deps() {
  echo "Checking dependencies..."
  nc -z localhost 55432 && echo "✓ PostgreSQL" || echo "✗ PostgreSQL"
  redis-cli -p 4050 ping > /dev/null 2>&1 && echo "✓ Redis" || echo "✗ Redis"
  curl -sf http://localhost:8200 > /dev/null && echo "✓ Vault" || echo "✗ Vault"
  # ... etc
}
```

## Common Tasks

### Reset Database

```bash
# Stop all services
./dev.sh stop

# Delete PostgreSQL volume
cd ../lma-devstack-compose-gitea4001
docker-compose down -v

# Restart and re-run migrations
./dev.sh start
cd ../services/projects-service
pnpm db:migrate:up
```

### Clear Redis Cache

```bash
redis-cli -p 4050 FLUSHALL
```

### Restart Single Service

```bash
# Find process ID
ps aux | grep "pnpm dev"

# Kill process
kill <pid>

# Restart
cd services/projects-service
pnpm dev
```

### View All Logs

```bash
# Infrastructure
cd ../lma-devstack-compose-gitea4001
docker-compose logs -f

# Specific infrastructure service
docker-compose logs -f postgres
docker-compose logs -f redis
```

## Troubleshooting

### Issue: Port Already in Use
```bash
# Find process using port
lsof -i :7002

# Kill process
kill -9 <pid>
```

### Issue: Service Won't Start
**Check dependencies:**
```bash
./dev.sh deps
```

**Check logs:**
```bash
cd services/<service-name>
cat dev.log  # If tee'd to file
```

### Issue: Database Connection Failed
**Verify PostgreSQL is running:**
```bash
docker-compose ps postgres
```

**Check connection:**
```bash
psql postgresql://lma:lma123@localhost:55432/lma_dev
```

### Issue: Hot-Reload Not Working
**Node (tsx):**
```bash
# Check tsx is watching
ps aux | grep tsx
```

**Go (air):**
```bash
# Check .air.toml exists
cat .air.toml

# Rebuild air
go install github.com/cosmtrek/air@latest
```

**Rust (cargo-watch):**
```bash
# Check cargo-watch is installed
cargo watch --version

# Reinstall
cargo install cargo-watch
```

## Best Practices

1. **Always start infrastructure first:**
   ```bash
   ./dev.sh start
   ```

2. **Use Hivemind for all services:**
   - Easier to manage multiple processes
   - Better log visibility
   - Consistent environment

3. **Keep .env.local in sync:**
   - Update when adding new services
   - Share template in `.env.example`

4. **Clean up regularly:**
   ```bash
   ./dev.sh clean  # Weekly or when switching branches
   ```

5. **Check health before testing:**
   ```bash
   curl http://localhost:7002/healthz
   ```

## Useful Links

- **Service Catalog:** `../service_catalog.yaml`
- **Root CLAUDE.md:** `../CLAUDE.md`
- **Infrastructure Setup:** `../lma-devstack-compose-gitea4001/README.md`
- **Hivemind Docs:** https://github.com/DarthSim/hivemind
