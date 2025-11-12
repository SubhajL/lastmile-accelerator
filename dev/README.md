# LMA Development Environment

Hybrid development setup for Sprint 1 services with **infrastructure in Docker** and **application services running natively** with hot-reload.

## 🎯 Philosophy

- **Infrastructure** (Postgres, Redis, Vault, etc.) → Docker Compose (consistent, isolated)
- **App Services** (Node, Go, Rust) → Native processes (fast hot-reload, easy debugging)
- **One command** to start everything
- **No external dependencies** (Gitea instead of GitHub, MailHog instead of real SMTP)

## 📦 Sprint 1 Services

| Service | Language | Port | Hot-Reload | Purpose |
|---------|----------|------|------------|---------|
| **projects-service** | Node.js | 7002 | `tsx watch` | IDs/tenants management |
| **observability-service** | Go | 7301 | `air` | Error inbox |
| **notification-service** | Node.js | 7902 | `tsx watch` | Email/Slack notifications |
| **dep-governance-service** | Rust | 7106 | `cargo-watch` | Dependency scanning |
| **db-guardian-service** | Go | 7105 | `air` | Database checks |
| **test-lab-service** | Node.js | 7202 | `tsx watch` | Testing orchestration |
| **secrets-env-service** | Go | 7104 | `air` | Vault integration |

## 🚀 Quick Start

### 1. Install Prerequisites

```bash
# Install dev tools
./dev.sh install

# Or manually:
brew install hivemind                          # Process manager
go install github.com/cosmtrek/air@latest      # Go hot-reload
cargo install cargo-watch                      # Rust hot-reload
```

### 2. Start Everything

```bash
cd dev
./dev.sh start
```

This will:
1. ✅ Start infrastructure (Docker Compose)
2. ✅ Wait for dependencies to be ready
3. ✅ Start all 7 services with hot-reload in tmux

### 3. Develop

Edit code in any service - changes reload automatically!

**Node.js** services reload in <1s  
**Go** services rebuild and restart in ~2s  
**Rust** rebuilds in 3-5s

### 4. Stop

```bash
# In another terminal
./dev.sh stop

# Or press Ctrl+C in the hivemind terminal
```

## 📖 Commands

```bash
./dev.sh start          # Start infrastructure + all services
./dev.sh stop           # Stop everything
./dev.sh restart        # Restart services (keeps infra running)
./dev.sh status         # Show what's running
./dev.sh logs           # Tail infrastructure logs
./dev.sh deps           # Check if infrastructure is ready
./dev.sh clean          # Stop and delete all volumes
./dev.sh install        # Install dev tools
```

## 🔧 Infrastructure URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / admin |
| **Prometheus** | http://localhost:9090 | - |
| **Gitea** | http://localhost:4001 | Create account on first visit |
| **MailHog** | http://localhost:8025 | - |
| **MinIO Console** | http://localhost:9001 | minioadmin / minioadmin |
| **Vault** | http://localhost:8200 | Token: `lma-root` |
| **Keycloak** | http://localhost:8080 | admin / admin |

## 🐛 Debugging

### Individual Service Logs

Hivemind shows all services in tmux panes. Navigate with:
- `Ctrl+b` then arrow keys → Switch panes
- `Ctrl+b` then `z` → Zoom current pane
- `Ctrl+b` then `d` → Detach (keep running in background)
- `tmux attach` → Reattach

### Run Single Service

```bash
cd ../services/projects-service
SERVICE_PORT=7002 SERVICE_NAME=projects-service pnpm dev
```

### Check Service Health

```bash
curl http://localhost:7002/healthz  # projects-service
curl http://localhost:7301/healthz  # observability-service
```

### View Infrastructure Logs

```bash
./dev.sh logs
# Or specific service:
cd ../lma-devstack-compose-gitea4001
docker compose logs -f postgres
```

## 🔐 Environment Variables

All configured in `.env.local`:
- Database → `DATABASE_URL=postgres://lma:lma@localhost:55432/lma`
- Redis → `REDIS_URL=redis://localhost:4050`
- MinIO → `S3_ENDPOINT=http://localhost:9000`
- Vault → `VAULT_ADDR=http://localhost:8200`
- OIDC → `OIDC_ISSUER_URL=http://localhost:8080/realms/lma`

See `.env.local` for complete list.

## 📁 File Structure

```
dev/
├── dev.sh                    # Main orchestrator
├── Procfile                  # Hivemind process definitions
├── .env.local                # Environment variables
├── wait-for-deps.sh         # Dependency checker
├── README.md                 # This file
└── hot-reload-configs/
    └── .air.toml            # Go hot-reload config
```

## 🎨 Hot-Reload Details

### Node.js (tsx watch)
- Watches `src/**/*.ts`
- Restarts on file changes
- Preserves console output
- **Fast**: <1s reload time

### Go (air)
- Watches `cmd/**/*.go` and `**/*.go`
- Rebuilds binary on changes
- Kills old process gracefully
- **Medium**: ~2s rebuild + restart

### Rust (cargo-watch)
- Watches `src/**/*.rs`
- Rebuilds on changes
- Shows compiler errors inline
- **Slower**: 3-5s rebuild (incremental compilation helps)

## 🛠️ Troubleshooting

### "hivemind not found"
```bash
brew install hivemind
```

### "air: command not found"
```bash
go install github.com/cosmtrek/air@latest
# Ensure $GOPATH/bin is in PATH
```

### "cargo-watch not found"
```bash
cargo install cargo-watch
```

### Infrastructure not starting
```bash
cd ../lma-devstack-compose-gitea4001
docker compose down -v  # Clean slate
docker compose up -d
./dev.sh deps           # Check readiness
```

### Port already in use
```bash
# Find what's using the port
lsof -i :7002

# Kill it
kill -9 <PID>
```

### Dependencies not ready
```bash
./dev.sh deps  # Diagnose
./dev.sh logs  # Check what's failing
```

## 🔄 Workflow Tips

### Test Email Sending
1. Send email from notification-service
2. View in MailHog: http://localhost:8025
3. No real email accounts needed!

### Test Git Operations
1. Use Gitea: http://localhost:4001
2. Create repos, push code
3. Test webhooks locally
4. No GitHub rate limits!

### View Metrics
1. Services send to OpenTelemetry: http://localhost:4318
2. Prometheus scrapes: http://localhost:9090
3. Visualize in Grafana: http://localhost:3000

### Database Access
```bash
# CLI
psql postgres://lma:lma@localhost:55432/lma

# GUI (use any Postgres client)
Host: localhost
Port: 55432
User: lma
Password: lma
Database: lma
```

## 🚦 CI/CD Integration

### Gitea Actions (Local)
1. Create `.gitea/workflows/` in your repo
2. Push to Gitea
3. Workflows run locally
4. No GitHub Actions minutes consumed!

### GitHub Actions (Production)
1. Use same workflow files (mostly compatible)
2. Push to GitHub for production builds
3. Best of both worlds

## 📚 Next Steps

1. ✅ Start infrastructure: `./dev.sh start`
2. ⏳ Edit code in services - watch them reload!
3. ⏳ View logs in Grafana
4. ⏳ Test features end-to-end
5. ⏳ Push to Gitea, set up webhooks

## 🆘 Getting Help

**Infrastructure issues?**
- Check logs: `./dev.sh logs`
- Restart: `cd ../lma-devstack-compose-gitea4001 && docker compose restart`

**Service not starting?**
- Check dependencies: `./dev.sh deps`
- Run individually to see errors (see "Run Single Service" above)

**Hot-reload not working?**
- Verify tool installed: `which air` / `which cargo-watch`
- Check file watchers: `ulimit -n` (should be >1024)

**Performance issues?**
- Docker resources: Increase CPU/memory in Docker Desktop settings
- Disable unused services in `Procfile` (comment out lines)

---

**Happy coding! 🚀**

For questions about specific services, see their `services/*/CONTEXT.md` files.
