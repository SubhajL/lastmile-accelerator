# SCM App Service - Manage source control management integrations

**Technology:** Go
**Ports:** REST: 7051, gRPC: 50041
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

Implements SCM app service responsibilities per PRD. Manages integrations with version control systems (Git, GitHub, GitLab), handles webhooks, and provides unified SCM operations.

## Quick Start

### Development
```bash
cd services/scm-app-service
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
scm-app-service/
├── cmd/scm/
│   └── main.go           # Entry point
├── internal/
│   ├── config/           # Configuration
│   ├── server/           # HTTP server
│   ├── grpc/             # gRPC server
│   ├── service/          # SCM logic
│   ├── clients/          # Git/GitHub/GitLab clients
│   ├── models/           # Domain models
│   └── telemetry/        # OpenTelemetry setup
├── api/                  # Protobuf definitions
├── Makefile
└── go.mod
```

## Key Files

- `cmd/scm/main.go` - Entry point
- `internal/server/server.go` - HTTP server setup
- `internal/clients/git.go` - Git operations
- `internal/service/scm.go` - SCM service logic

## API

**REST Endpoints:**
- `GET /healthz` - Health check
- `GET /metrics` - Metrics (Prometheus format)
- `POST /repos/list` - List repositories
- `POST /repos/{repoId}/pull-request` - Create PR

**gRPC:** See `api/scm.proto`

## Configuration

Key environment variables:
- `SERVICE_NAME=scm-app-service`
- `SERVICE_PORT=7051`
- `GRPC_PORT=50041`
- `GITHUB_TOKEN` - GitHub API token
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector
- `VAULT_ADDR`, `VAULT_ROLE_ID`, `VAULT_SECRET_ID` - Vault secrets

See [../CLAUDE.md](../CLAUDE.md) for standard service configuration.

## Testing

- **Unit:** `internal/**/*_test.go`
- **Run:** `go test ./...`
- **Coverage:** `go tool cover -html=coverage.out`

## Common Patterns

- **Webhook handling:** Process GitHub/GitLab webhooks
- **Multi-provider support:** Abstract over different SCM providers
- **Token management:** Secure credential storage in Vault

## Related Services

- **projects-service:** Stores project-SCM mappings
- **notification-service:** Sends SCM event notifications

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: `./CONTEXT.md`
- CI Workflow: `../../.github/workflows/ci-scm-app-service.yml`
