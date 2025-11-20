# Projects Service - Project Lifecycle Management

**Technology:** Node.js/TypeScript, Fastify
**Ports:** REST: 7002, gRPC: 50052
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Projects Service is the central hub for project lifecycle management in the LMA platform. It manages project metadata (name, description, team, tech stack), orchestrates project creation workflow across services, tracks project status and milestones, enforces project-level permissions and quotas, and provides unified project dashboard data aggregation. Every other service references projects to scope their operations.

## Development Commands

### From This Directory
```bash
# Node service commands
pnpm install          # Install dependencies
pnpm dev              # Hot-reload with tsx watch
pnpm test             # Run tests with Vitest
pnpm test:ui          # Vitest UI for interactive testing
pnpm test:coverage    # Generate coverage report
pnpm ci:test          # CI-optimized test run
pnpm ci:coverage      # CI coverage report
pnpm typecheck        # Type checking without building
pnpm lint             # ESLint with max-warnings=0
pnpm build            # Build for production
pnpm start            # Run production build

# Database operations
pnpm migrate          # Run database migrations
pnpm seed             # Seed database with sample data

# Utilities
pnpm kill:watchers    # Kill orphaned file watchers
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=projects-service
bunx turbo run test --filter=projects-service
bunx turbo run build --filter=projects-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
pnpm typecheck && pnpm lint && pnpm test && pnpm build

# Or from root with turbo
bunx turbo run typecheck lint test build --filter=projects-service
```

## Architecture

### Directory Structure
```
projects-service/
├── src/
│   ├── __tests__/               # Test suites
│   │   ├── projects.test.ts
│   │   ├── teams.test.ts
│   │   └── integration/
│   ├── db/                      # Database layer
│   │   ├── client.ts           # PostgreSQL connection
│   │   ├── migrations/         # SQL migration files
│   │   │   └── migrate.ts      # Migration runner
│   │   ├── repositories/
│   │   │   ├── projects.repo.ts
│   │   │   ├── teams.repo.ts
│   │   │   └── quotas.repo.ts
│   │   └── seed.ts             # Database seeding
│   ├── events/                  # NATS event handling
│   │   ├── publisher.ts
│   │   └── subscribers/
│   │       ├── user-created.ts
│   │       └── deployment-completed.ts
│   ├── middleware/              # Fastify middleware
│   │   ├── auth.middleware.ts
│   │   ├── error.middleware.ts
│   │   └── rate-limit.middleware.ts
│   ├── observability/           # Telemetry setup
│   │   ├── metrics.ts
│   │   └── tracing.ts
│   ├── routes/                  # REST endpoint handlers
│   │   ├── projects.routes.ts
│   │   ├── teams.routes.ts
│   │   ├── quotas.routes.ts
│   │   └── health.routes.ts
│   ├── services/                # Business logic
│   │   ├── projects.service.ts
│   │   ├── teams.service.ts
│   │   ├── quotas.service.ts
│   │   └── orchestrator.service.ts
│   ├── utils/                   # Utilities
│   │   ├── logger.ts
│   │   ├── errors.ts
│   │   └── validation.ts
│   ├── app.ts                   # Fastify app setup
│   ├── config.ts                # Configuration management
│   ├── index.ts                 # Entry point
│   └── types.ts                 # TypeScript types
├── scripts/
│   └── kill-watchers.sh         # Cleanup script
├── docs/                        # Documentation
│   ├── api-spec.yaml           # OpenAPI specification
│   └── architecture.md
├── coverage/                    # Test coverage reports
├── dist/                        # Build output
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── vitest.config.ts             # Test configuration
├── Makefile                     # Build shortcuts
├── BUILD_PLAN.md                # Development plan
├── IMPLEMENTATION_STATUS.md     # Build progress
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Project Orchestrator:**
- File: `src/services/orchestrator.service.ts` - Coordinates project creation workflow
- Pattern: Publishes events to initialize project in all dependent services
- Example: On project creation, triggers SBOM generation, test setup, notification preferences

**Project Repository:**
- File: `src/db/repositories/projects.repo.ts` - Project CRUD operations
- Pattern: PostgreSQL queries with parameterized statements
- Example: Stores project metadata with team associations

**Team Management:**
- File: `src/services/teams.service.ts` - Manages project teams and permissions
- Pattern: Role-based access control (owner, admin, developer, viewer)
- Example: Validates user permissions before allowing project modifications

**Quota Enforcement:**
- File: `src/services/quotas.service.ts` - Enforces resource limits per project
- Pattern: Checks quotas before resource creation (max deployments, storage, test runs)
- Example: Blocks new deployment if project exceeds monthly deployment quota

**Migration Runner:**
- File: `src/db/migrations/migrate.ts` - Custom migration framework
- Pattern: Versioned SQL migrations with up/down support
- Example: Runs pending migrations on service startup

### Dependencies

**Core:**
- `fastify` 4.27.0 - Fast web framework for REST API
- `@fastify/cors` ^8.4.2 - CORS support
- `@fastify/helmet` ^11.1.1 - Security headers
- `pg` ^8.11.3 - PostgreSQL client
- `nats` ^2.25.0 - NATS messaging for event-driven workflows
- `zod` ^3.22.4 - Schema validation
- `uuid` ^9.0.1 - UUID generation
- `pino` ^8.17.2 - Structured logging
- `jose` ^5.2.3 - JWT handling (modern alternative to jsonwebtoken)
- `jsonwebtoken` ^9.0.2 - JWT authentication (legacy support)

**Observability:**
- `@opentelemetry/api` ^1.9.0 - OpenTelemetry tracing
- `@opentelemetry/sdk-node` ^0.55.0 - OTel SDK
- `@opentelemetry/auto-instrumentations-node` ^0.55.0 - Auto instrumentation
- `@opentelemetry/exporter-trace-otlp-http` ^0.55.0 - OTLP trace exporter
- `@opentelemetry/exporter-metrics-otlp-http` ^0.55.0 - OTLP metrics exporter
- `@opentelemetry/resources` ^1.9.0 - Resource attributes
- `@opentelemetry/semantic-conventions` ^1.26.0 - Standard attributes

**Testing:**
- `vitest` ^1.1.0 - Fast unit test framework
- `@vitest/ui` ^1.1.0 - Interactive test UI
- `@vitest/coverage-v8` ^1.6.1 - Code coverage
- `supertest` ^6.3.3 - HTTP assertions
- `@testcontainers/postgresql` ^11.8.0 - PostgreSQL Testcontainers
- `testcontainers` ^11.8.0 - Container-based integration tests

**Development:**
- `tsx` 4.19.2 - TypeScript execution with watch mode
- `typescript` 5.6.3 - TypeScript compiler
- `pino-pretty` ^10.3.1 - Pretty logs for development
- `@apidevtools/swagger-parser` ^12.1.0 - OpenAPI validation

## Code Organization Patterns

### Request Handlers
✅ **DO:** Use route -> service -> repository pattern
```typescript
// src/routes/projects.routes.ts
export async function projectsRoutes(fastify: FastifyInstance) {
  fastify.post('/projects', async (request, reply) => {
    const project = await projectsService.createProject(request.body);
    return reply.code(201).send(project);
  });
}
```
❌ **DON'T:** Put business logic in route handlers

### Service Layer
✅ **DO:** Implement business logic and orchestration in services
```typescript
// src/services/projects.service.ts
export class ProjectsService {
  async createProject(data: CreateProjectDTO): Promise<Project> {
    const project = await this.repo.create(data);
    await this.orchestrator.initializeProject(project.id);
    await this.events.publish('project.created', project);
    return project;
  }
}
```
❌ **DON'T:** Skip event publishing or orchestration steps

### Database Access
✅ **DO:** Use repository pattern with async/await
```typescript
// src/db/repositories/projects.repo.ts
export async function createProject(project: NewProject): Promise<Project> {
  const result = await pool.query(
    'INSERT INTO projects (name, owner_id, tech_stack) VALUES ($1, $2, $3) RETURNING *',
    [project.name, project.ownerId, JSON.stringify(project.techStack)]
  );
  return result.rows[0];
}
```
❌ **DON'T:** Write SQL in service layer

### Error Handling
✅ **DO:** Use typed errors and centralized error middleware
```typescript
// src/utils/errors.ts
export class ProjectNotFoundError extends Error {
  constructor(projectId: string) {
    super(`Project ${projectId} not found`);
    this.name = 'ProjectNotFoundError';
  }
}
```
❌ **DON'T:** Return inconsistent error formats

## API Endpoints

### REST API

**Base URL:** `http://localhost:7002`

**Key Endpoints:**
- `GET /healthz` - Health check with dependency status
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/projects` - Create new project
- `GET /api/v1/projects` - List projects (with pagination)
- `GET /api/v1/projects/{id}` - Get project details
- `PUT /api/v1/projects/{id}` - Update project metadata
- `DELETE /api/v1/projects/{id}` - Soft delete project
- `POST /api/v1/projects/{id}/teams` - Add team member
- `DELETE /api/v1/projects/{id}/teams/{userId}` - Remove team member
- `GET /api/v1/projects/{id}/quotas` - Get project resource quotas
- `PUT /api/v1/projects/{id}/quotas` - Update quotas
- `GET /api/v1/projects/{id}/dashboard` - Get aggregated dashboard data

### gRPC API

**Port:** 50052

**Services:**
- `ProjectsService` - Project lifecycle management
  - `CreateProject` - Create new project
  - `GetProject` - Retrieve project details
  - `ListProjects` - List projects with filters
  - `UpdateProject` - Modify project metadata
  - `DeleteProject` - Soft delete project

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`projects`** - Core project metadata
- Columns: `id`, `name`, `description`, `owner_id`, `tech_stack` (JSONB), `status`, `created_at`, `updated_at`, `deleted_at`
- Indexes: `idx_owner_id`, `idx_status`, `idx_created_at`
- Purpose: Store project information and configuration

**`project_teams`** - Team membership
- Columns: `project_id`, `user_id`, `role` (owner|admin|developer|viewer), `joined_at`
- Indexes: `idx_project_id`, `idx_user_id`, `unique_idx_project_user`
- Purpose: Track who has access to each project

**`project_quotas`** - Resource limits
- Columns: `project_id`, `max_deployments_per_month`, `max_storage_gb`, `max_test_runs_per_day`, `current_usage` (JSONB)
- Indexes: `idx_project_id`
- Purpose: Enforce resource quotas per project

**`project_milestones`** - Project milestones
- Columns: `id`, `project_id`, `name`, `description`, `due_date`, `completed_at`, `created_at`
- Indexes: `idx_project_id`, `idx_due_date`
- Purpose: Track project progress and deadlines

**Migrations:**
- Location: `src/db/migrations/`
- Tool: Custom migration scripts
- Commands: `pnpm migrate` (runs `src/db/migrations/migrate.ts`)

## Event Handling

**Published Events:**
- `project.created` - When new project created
  - Payload: `{project_id, name, owner_id, tech_stack, created_at}`
- `project.updated` - When project metadata changed
  - Payload: `{project_id, changed_fields, updated_by, updated_at}`
- `project.deleted` - When project soft deleted
  - Payload: `{project_id, deleted_by, deleted_at}`
- `project.team.added` - When team member added
  - Payload: `{project_id, user_id, role, added_by}`
- `project.quota.exceeded` - When resource quota exceeded
  - Payload: `{project_id, quota_type, limit, current_usage}`

**Subscribed Events:**
- `user.created` - Initialize user's personal project workspace
- `deployment.completed` - Update project deployment count
- `test-run.completed` - Update project test run count

## Testing Strategy

### Unit Tests
- Location: `src/__tests__/*.test.ts`
- Coverage: Target >80%
- Mock: Database with in-memory PostgreSQL or mocks, NATS, external services
- Example: `src/__tests__/projects.test.ts`

### Integration Tests
- Location: `src/__tests__/integration/`
- Setup: Use Testcontainers for PostgreSQL
- Pattern: Test full request flow from HTTP -> service -> database

### Running Tests
```bash
# All tests
pnpm test

# Watch mode
pnpm test:watch

# With UI
pnpm test:ui

# Coverage report
pnpm test:coverage

# CI-optimized (no watch, single run)
pnpm ci:test
```

## Configuration

### Environment Variables
```bash
# Service-specific config
DEFAULT_PROJECT_QUOTA_DEPLOYMENTS=100
DEFAULT_PROJECT_QUOTA_STORAGE_GB=10
DEFAULT_PROJECT_QUOTA_TEST_RUNS=1000
ENABLE_PROJECT_DELETION=false  # Soft delete only
MAX_PROJECTS_PER_USER=50

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=projects-service
SERVICE_PORT=7002
GRPC_PORT=50052
ENV=dev
DATABASE_URL=postgresql://user:pass@localhost:5432/projects
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/projects-service/`
- Accessed via: `secrets-env-service` or direct Vault SDK
- Keys: Database credentials, API keys for external integrations

## Quick Find Commands

### Find Code
```bash
# Find project creation logic
rg -n "createProject|create.*project" services/projects-service/src/

# Find database queries
rg -n "pool\.query|db\.query" services/projects-service/src/db/

# Find event publishers
rg -n "publish.*project\." services/projects-service/src/events/

# Find route definitions
rg -n "fastify\.(get|post|put|delete)" services/projects-service/src/routes/

# Find team management
rg -n "addTeamMember|removeTeamMember|teams" services/projects-service/src/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "projects-service|localhost:7002" --glob "docker-compose*.yml" --glob "*.yaml"

# Find all services querying project data
rg -n "projects.*getProject|fetch.*project" services/
```

## Common Gotchas

- **Soft Delete Cascades:** Deleting a project should cascade to dependent resources (teams, quotas, milestones); ensure cascading deletes or cleanup jobs
- **Quota Race Conditions:** Concurrent requests may exceed quotas; use database-level constraints or pessimistic locking
- **Team Permissions:** Validate user permissions before all project operations; missing checks create security holes
- **Event Ordering:** Project creation events must complete before dependent services initialize; use event acknowledgments
- **Migration Failures:** Failed migrations leave database in inconsistent state; always test migrations on copy of production data
- **JSON Column Queries:** Querying JSONB `tech_stack` column is slow without GIN index; ensure proper indexing
- **Pagination Limits:** Large projects lists can timeout; always paginate with reasonable limits (e.g., 50 items per page)

## Related Services

- **db-guardian-service:** Validates database migrations for projects; Projects Service triggers validation
- **test-lab-service:** Executes tests for projects; Projects Service provides project configuration
- **dep-governance-service:** Tracks dependencies for projects; Projects Service links SBOMs to projects
- **notification-service:** Sends project alerts; Projects Service publishes project events
- **secrets-env-service:** Manages project secrets; Projects Service references secret paths
- **observability-service:** Aggregates project metrics; Projects Service provides project context for filtering

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-projects-service.yml`
- Build Plan: `BUILD_PLAN.md`
- Implementation Status: `IMPLEMENTATION_STATUS.md`
- Architecture Phases: `PHASE_1_2_PLAN.md`, `PHASE_1_2_SUMMARY.md`
