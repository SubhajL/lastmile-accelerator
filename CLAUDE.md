# Last-Mile Accelerator (LMA)

## Overview
- **Type:** Polyglot Monorepo (Turborepo)
- **Stack:** Node.js, Go, Rust, Python, Next.js
- **Architecture:** 28 microservices + 3 frontends + shared infrastructure
- **Purpose:** Ship-ready guardrails for developers—security, tests, telemetry, and deploys that just work

This CLAUDE.md is the authoritative source for development guidelines.
Subdirectories contain specialized CLAUDE.md files that extend these rules.

## Universal Development Rules

### Code Quality (MUST)
- **MUST** follow language-specific type systems strictly (TypeScript strict mode, Go type safety, Rust ownership)
- **MUST** write tests for all new features (TDD: scaffold stub → failing test → implement)
- **MUST** run pre-commit hooks before committing (formatting, linting, security scanning)
- **MUST NOT** commit secrets, API keys, tokens, or credentials (enforced by gitleaks)
- **MUST** separate pure-logic unit tests from DB-touching integration tests
- **MUST** pass all CI gates: typecheck, lint, test, build, security scans
- **MUST** use conventional commits: `<type>[scope]: <description>` (feat, fix, docs, refactor, test, chore)

### Best Practices (SHOULD)
- **SHOULD** prefer functional, composable, testable code over complex classes
- **SHOULD** keep functions under 50 lines; extract complex logic into separate functions
- **SHOULD** use descriptive variable names (avoid single letters except loop counters)
- **SHOULD** colocate unit tests with source files (`*.spec.ts`, `*_test.go`)
- **SHOULD** prefer integration tests over heavy mocking for database operations
- **SHOULD** use branded types for IDs in TypeScript: `type UserId = Brand<string, 'UserId'>`
- **SHOULD** use `import type { ... }` for type-only imports in TypeScript
- **SHOULD** default to `type` over `interface` unless merging is required

### Anti-Patterns (MUST NOT)
- **MUST NOT** use `any` type without explicit justification (or Go `interface{}` without reason)
- **MUST NOT** bypass type errors with `@ts-ignore` or `@ts-expect-error`
- **MUST NOT** push directly to main branch (use PRs with reviews)
- **MUST NOT** disable security tooling (gitleaks, hadolint, OPA policies)
- **MUST NOT** use `:latest` tag in Dockerfiles (pin specific versions)
- **MUST NOT** run containers as root (use distroless or non-root users)
- **MUST NOT** ignore errors in Go or Result types in Rust
- **MUST NOT** hardcode environment-specific values (use config/env vars)

## Core Commands

### Monorepo Management
```bash
# Install all dependencies (root + workspaces)
bun install

# Run command across all packages
bunx turbo run <command>

# Run command for specific package
bunx turbo run <command> --filter=<package-name>

# Run command for changed packages only
bunx turbo run <command> --filter='...[HEAD^]'
```

### Development
```bash
# Start local dev stack (infra + 7 Sprint 1 services)
cd dev && ./dev.sh start

# Check dev stack status
cd dev && ./dev.sh status

# View infrastructure logs
cd dev && ./dev.sh logs

# Restart services (keeps infrastructure running)
cd dev && ./dev.sh restart

# Stop everything
cd dev && ./dev.sh stop

# Clean all volumes and data
cd dev && ./dev.sh clean
```

### Quality Gates (run before PR)
```bash
# Root level - all packages
bunx turbo run typecheck lint test build

# Specific service (Node/TypeScript)
cd services/<service-name>
pnpm typecheck && pnpm test && pnpm build

# Specific service (Go)
cd services/<service-name>
make test && make build

# Specific service (Rust)
cd services/<service-name>
cargo test --all --locked && cargo build --release

# Specific service (Python)
cd services/<service-name>
pytest -q
```

### Git Worktree Management
```bash
# List all worktrees
git worktree list

# Add new worktree for a service
git worktree add ~/dev/worktrees/<service-name> -b <branch-name>

# Remove worktree
git worktree remove ~/dev/worktrees/<service-name>

# Prune deleted worktrees
git worktree prune
```

## Project Structure

### Services (28 microservices)
**Location:** `services/`
**Documentation:** Each service has `CLAUDE.md`, `AGENTS.md`, `CONTEXT.md`

**Service Categories:**
- **Core Infrastructure** (7 services - Sprint 1 / Active in Worktrees)
  - `projects-service` → Node (REST: 7002, gRPC: 50052) - [CLAUDE.md](services/projects-service/CLAUDE.md)
  - `observability-service` → Go (REST: 7301, gRPC: 50081) - [CLAUDE.md](services/observability-service/CLAUDE.md)
  - `notification-service` → Node (REST: 7902, gRPC: 50122) - [CLAUDE.md](services/notification-service/CLAUDE.md)
  - `dep-governance-service` → Rust (REST: 7106, gRPC: 50066) - [CLAUDE.md](services/dep-governance-service/CLAUDE.md)
  - `db-guardian-service` → Go (REST: 7105, gRPC: 50065) - [CLAUDE.md](services/db-guardian-service/CLAUDE.md)
  - `test-lab-service` → Node (REST: 7202, gRPC: 50072) - [CLAUDE.md](services/test-lab-service/CLAUDE.md)
  - `secrets-env-service` → Go (REST: 7104, gRPC: 50064) - [CLAUDE.md](services/secrets-env-service/CLAUDE.md)

**See:** `service_catalog.yaml` for complete list with ports

### Frontends (3 Next.js applications)
**Location:** `frontends/`
**Package Manager:** pnpm (NOT bun)
**Framework:** Next.js 14.2.6 (App Router)

- `mode-a-dashboard` - [CLAUDE.md](frontends/mode-a-dashboard/CLAUDE.md)
- `mode-b-privacy-portal` - [CLAUDE.md](frontends/mode-b-privacy-portal/CLAUDE.md)
- `mode-c-zip-uploader` - [CLAUDE.md](frontends/mode-c-zip-uploader/CLAUDE.md)

### Infrastructure
- **`infra/`** → Kubernetes, Helm charts, Envoy configs - [CLAUDE.md](infra/CLAUDE.md)
- **`.github/workflows/`** → CI/CD pipelines - [CLAUDE.md](.github/CLAUDE.md)
- **`dev/`** → Local development orchestration - [CLAUDE.md](dev/CLAUDE.md)
- **`lma-devstack-compose-gitea4001/`** → Docker Compose dev stack

### Testing
- **Unit tests:** Colocated with source
  - Node: `*.spec.ts` (Vitest)
  - Go: `*_test.go` (go test)
  - Rust: `tests/` directory (cargo test)
  - Python: `test_*.py` (pytest)
- **Integration:** Service-specific `tests/integration/`
- **E2E:** Frontend `tests/e2e/` (Playwright)

## Quick Find Commands

### Service Navigation
```bash
# Find service by port
rg -n "REST.*7002" service_catalog.yaml

# Find all services by language
rg -n "lang: node" service_catalog.yaml
rg -n "lang: go" service_catalog.yaml

# List all service directories
ls -1 services/
```

### Code Search
```bash
# Find component/function definition (Node)
rg -n "export (function|const) .*MyFunction" services/

# Find API endpoint handlers (Node/Fastify)
rg -n "fastify\.(get|post|put|delete)" services/

# Find gRPC service definitions
rg -n "service.*\{" --glob "*.proto"

# Find type definitions (TypeScript)
rg -n "^export (type|interface)" services/ packages/

# Find hook usage (React)
rg -n "use[A-Z]" frontends/

# Find database queries (Go)
rg -n "db\.Query|db\.Exec" services/

# Find Rust async functions
rg -n "async fn" services/dep-governance-service/
```

### Dependency Analysis
```bash
# Check package dependencies
bun pm ls <package-name>

# Find which services use a specific package (Node)
rg -n "\"<package-name>\"" --glob "package.json"

# Check Go module dependencies
cd services/<go-service> && go mod graph | grep <module>

# Find unused dependencies (Node)
cd services/<node-service> && npx depcheck
```

### Git Operations
```bash
# Find recent commits for a service
git log --oneline --follow -- services/<service-name>/

# Find which branch a worktree uses
git worktree list

# Check if a branch exists remotely
git ls-remote --heads origin <branch-name>
```

## Security Guidelines

### Secrets Management
- **NEVER** commit tokens, API keys, or credentials to git
- Use `.env.local` for local secrets (already in .gitignore)
- Use Vault for production secrets (via `secrets-env-service`)
- Use GitHub Actions secrets for CI/CD tokens
- **ALWAYS** run gitleaks before commits (enforced by CI)
- Redact PII in logs and error messages

### Allowed Environment Variables
```bash
# Development (local only)
DATABASE_URL=postgresql://localhost:55432/lma_dev
REDIS_URL=redis://localhost:4050
VAULT_ADDR=http://localhost:8200

# Production (via Vault or K8s secrets)
# NEVER hardcode production credentials
```

### Safe Operations
- Review generated bash commands before execution
- Require confirmation before:
  - `git push --force`
  - `rm -rf` commands
  - Database migrations in production
  - `kubectl delete` operations
  - Docker image publishes
- Use staging environment for risky operations
- Always have rollback plan for migrations

### Security Scanning (Automated)
- **Gitleaks:** Scans for secrets in commits (`.github/gitleaks.toml`)
- **Hadolint:** Lints Dockerfiles for best practices
- **OPA Conftest:** Policy checks on K8s YAML (`.github/policy/root.rego`)
- **Syft:** Generates SBOM for container images
- **Grype:** Scans container images for CVEs

## Git Workflow

### Branch Strategy
**Main Branch:** `main` (standardized from `db-guardian-service`)
**Branch Naming:**
- Feature: `feat/<description>` or `<service>/<feature>`
- Bug fix: `fix/<description>`
- Documentation: `docs/<description>`
- Chores: `chore/<description>`
- Service-specific: `<service-name>` (for major service work)

### Worktree Workflow
**Active Worktrees:** 7 services in `~/dev/worktrees/`
- Each worktree is a separate git checkout on its own branch
- Enables parallel development without branch switching
- Keeps build artifacts and node_modules isolated
- Use custom slash commands: `/worktree-switch <service>`

### Commit Messages
**Format:** Conventional Commits
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style (formatting, no logic change)
- `refactor:` - Code restructuring (no behavior change)
- `perf:` - Performance improvements
- `test:` - Adding or updating tests
- `build:` - Build system or dependencies
- `ci:` - CI/CD configuration
- `chore:` - Maintenance tasks

**Examples:**
```
feat(projects-service): add multi-tenancy support
fix(db-guardian): prevent migration race condition
docs: update worktree setup guide
test(notification): add email template rendering tests
```

### Pull Request Requirements
- ✅ All tests pass (unit + integration)
- ✅ Type checking passes
- ✅ Linting passes
- ✅ Gitleaks scan passes
- ✅ Hadolint passes (if Dockerfile changed)
- ✅ OPA policy checks pass (if K8s YAML changed)
- ✅ Build succeeds
- ✅ At least 1 approval from code owner
- ✅ Conventional commit messages

**PR Template Sections:**
1. **Summary:** What changed and why
2. **Test Plan:** How you verified the changes
3. **Security Considerations:** Any security implications
4. **Breaking Changes:** Migration path if applicable
5. **Screenshots:** For UI changes

### Merge Strategy
- **Squash and merge** for feature branches
- **Linear history** preferred
- Delete branch after merge
- Update CHANGELOG.md for significant features

## Testing Requirements

### Testing Philosophy
Follow Sabrina Ramonov's AI-Assisted Programming Guidelines (from root `AGENTS.md`):
- **T-1 (MUST):** Colocate unit tests: `*.spec.ts`, `*_test.go`, `test_*.py`
- **T-2 (MUST):** Add/extend integration tests for API changes
- **T-3 (MUST):** Separate pure-logic unit tests from DB integration tests
- **T-4 (SHOULD):** Prefer integration tests over heavy mocking
- **T-5 (SHOULD):** Unit-test complex algorithms thoroughly
- **T-6 (SHOULD):** Test entire structure in one assertion when possible

### Coverage Requirements
- **Unit tests:** All business logic (aim for >80% coverage)
- **Integration tests:** API endpoints, database operations, external service calls
- **E2E tests:** Critical user paths (frontends only)
- **Smoke tests:** Health checks, basic service startup

### Test Commands
```bash
# Node services (Vitest)
pnpm test                    # Run all tests
pnpm test:watch              # Watch mode
pnpm test:coverage           # Generate coverage report
pnpm test <file-path>        # Run specific test file

# Go services
make test                    # Run all tests with coverage
go test ./internal/...       # Run package tests
go test -v -run TestFoo      # Run specific test

# Rust services
cargo test --all --locked    # Run all tests
cargo test test_name         # Run specific test
cargo test -- --nocapture    # Show println! output

# Python services
pytest -q                    # Quiet mode
pytest -v test_module.py     # Verbose specific file
pytest --cov=src             # With coverage
```

### Database Testing
- Use **Testcontainers** for Node services (real PostgreSQL in Docker)
- Use **sqlmock** for Go services (lightweight mocking)
- Use **pg-mem** for in-memory PostgreSQL (Node unit tests)
- Always use transactions in tests; rollback after each test
- Seed test data in `beforeEach`/`setUp`

## Available Tools

### Standard Tools
You have access to:
- **Bash tools:** git, rg, curl, jq, yq, docker, kubectl
- **Node tools:** bun, pnpm, node, npx
- **Go tools:** go, goose (migrations)
- **Rust tools:** cargo, rustc
- **Python tools:** python3, pip, pytest
- **GitHub CLI:** `gh` for issues, PRs, releases
- **Dev stack:** Docker Compose (PostgreSQL, Redis, Vault, MinIO, NATS, Keycloak)

### MCP Servers (Model Context Protocol)
Configured MCP servers provide extended capabilities:
- **GitHub MCP:** Issues, PRs, repository management
- **Sequential Thinking MCP:** Complex architectural decisions
- **Context7 MCP:** Documentation search and retrieval
- **Postgres MCP:** Database schema inspection

### Tool Permissions

**Allowed (✅):**
- Read any file in repository
- Write code files (`*.ts`, `*.go`, `*.rs`, `*.py`, `*.tsx`)
- Write config files (`*.yaml`, `*.json`, `*.toml`, `Dockerfile`, `Makefile`)
- Run tests, linters, type checkers
- Run `git status`, `git diff`, `git log`
- Run `gh pr view`, `gh issue view`
- Build Docker images locally
- Run local development stack

**Require Permission (⚠️):**
- Edit `.env` files (ask first)
- Edit security configs (`.github/gitleaks.toml`, `policy/root.rego`)
- Run database migrations (non-local environments)
- Push Docker images to GHCR
- Create/merge PRs
- Force push to any branch
- Delete resources (databases, K8s namespaces)

**Blocked (❌):**
- Edit files in `.git/` directory
- Disable security hooks or scanning
- Commit secrets/credentials
- Run `rm -rf /` or similar destructive commands
- Bypass CI gates
- Push to `main` branch directly

## Specialized Context

When working in specific directories, refer to their CLAUDE.md for detailed guidance:

### Services
- **General service patterns:** [services/CLAUDE.md](services/CLAUDE.md)
- **Node services:** Examples in [services/projects-service/CLAUDE.md](services/projects-service/CLAUDE.md)
- **Go services:** Examples in [services/db-guardian-service/CLAUDE.md](services/db-guardian-service/CLAUDE.md)
- **Rust services:** See [services/dep-governance-service/CLAUDE.md](services/dep-governance-service/CLAUDE.md)
- **Python services:** See [services/ai-debugger-service/CLAUDE.md](services/ai-debugger-service/CLAUDE.md)

### Frontends
- **Next.js patterns:** [frontends/CLAUDE.md](frontends/CLAUDE.md)
- **Specific apps:** `frontends/<app-name>/CLAUDE.md`

### Infrastructure
- **CI/CD guidelines:** [.github/CLAUDE.md](.github/CLAUDE.md)
- **Kubernetes/Helm:** [infra/CLAUDE.md](infra/CLAUDE.md)
- **Local development:** [dev/CLAUDE.md](dev/CLAUDE.md)

## Custom Slash Commands

Use these commands for common workflows:

### Worktree Management
- `/worktree-list` - Show all worktrees and their branches
- `/worktree-switch <service>` - Navigate to service worktree
- `/worktree-sync` - Sync all worktrees with remote

### Testing Workflows
- `/test-service <service>` - Run tests for specific service
- `/test-all-services` - Run tests across all services
- `/test-affected` - Run tests for changed services only

### PR/Issue Workflows
- `/fix-issue <issue-number>` - Fetch issue, implement fix, create PR
- `/review-pr <pr-number>` - Comprehensive security + quality review
- `/pr-create <title>` - Create PR with standard template

### Quality Checks
- `/quality-check` - Run typecheck+lint+test+build
- `/security-scan` - Run gitleaks, hadolint, OPA policies
- `/sbom-generate <service>` - Generate SBOM for service

## Common Gotchas

### Monorepo Issues
- **Node services use Bun**, frontends use pnpm - don't mix them
- **Turborepo cache** can be stale - use `bunx turbo run <cmd> --force` to bypass
- **Worktrees share git objects** - commits in one worktree affect all
- **Each worktree has own node_modules** - reinstall if switching

### Service Communication
- **gRPC requires proto files** in `api/` directory - regenerate with `make proto`
- **NATS events are async** - use request-reply for synchronous needs
- **JWT tokens expire** - refresh from Keycloak (localhost:8080)
- **Service mesh mTLS** - use Envoy sidecar in production, plain HTTP in dev

### Database
- **Migrations are sequential** - never reorder or modify merged migrations
- **Use transactions** for multi-step operations
- **Connection pooling** - max 10 connections per service in dev
- **Timezone is UTC** - always use UTC in database, convert in application layer

### Docker
- **Distroless images** have no shell - can't exec into them (use debug variants)
- **Multi-stage builds** - source code not in final image
- **Volume mounts in dev** - changes reflect immediately (hot-reload)
- **Port conflicts** - check `service_catalog.yaml` for port assignments

### CI/CD
- **Gitleaks fails on test files** - use allowlisting in `.github/gitleaks.toml`
- **Hadolint requires LABEL** - add maintainer label to Dockerfiles
- **OPA policies block :latest** - pin specific image versions
- **Grype scans take time** - run locally with `grype <image>` before pushing

## Development Shortcuts (from AGENTS.md)

AI-assisted development shortcuts for common tasks:

- **QNEW** - Understand best practices before starting new work
- **QPLAN** - Analyze requirements and plan implementation approach
- **QCODE** - Implement feature following TDD (stub → test → code)
- **QCHECK** - Skeptical code review (correctness, performance, security)
- **QCHECKF** - Fast shallow review (syntax, obvious issues only)
- **QCHECKT** - Test-focused review (coverage, edge cases, assertions)
- **QUX** - Generate UX testing scenarios
- **QGIT** - Stage, commit, push with conventional commit message

## Documentation

### Centralized Documentation
**Location:** `documentation/`
```
documentation/
├── features/
│   ├── active/          # In-progress feature specs
│   ├── completed/       # Shipped features
│   └── planned/         # Future roadmap
└── architecture/        # System design docs
```

### Service Documentation
Each service contains:
- **CLAUDE.md** - Claude Code development guidelines (this file system)
- **AGENTS.md** - AI-assisted development patterns and shortcuts
- **CONTEXT.md** - Service purpose, ports, dependencies, SLOs, ownership
- **README.md** - Quick start and basic information

### Update Documentation
- Update `CONTEXT.md` when changing service configuration
- Update `AGENTS.md` when adding new patterns or shortcuts
- Update feature specs in `documentation/features/active/` during development
- Move feature specs to `completed/` when shipped

## Support

### Getting Help
- **Claude Code help:** `/help` command or https://github.com/anthropics/claude-code/issues
- **Project issues:** https://github.com/SubhajL/lastmile-accelerator/issues
- **Service-specific:** Check service's `AGENTS.md` and `CONTEXT.md`

### Useful Links
- **Service Catalog:** `service_catalog.yaml` (all services with ports)
- **CI Workflows:** `.github/workflows/`
- **OPA Policies:** `.github/policy/root.rego`
- **Gitleaks Config:** `.github/gitleaks.toml`
- **Dev Stack Guide:** `dev/README.md`

---

**Last Updated:** 2025-11-20
**Claude Code Version:** Optimized for Claude Code with hierarchical CLAUDE.md system
