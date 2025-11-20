# Test Lab Service - AI-Powered Testing & Debugging Automation

**Technology:** Node.js/TypeScript, Fastify
**Ports:** REST: 7202, gRPC: 50072
**Parent Context:** This extends [../../CLAUDE.md](../../CLAUDE.md) and [../CLAUDE.md](../CLAUDE.md)

## Service Purpose

The Test Lab Service provides AI-powered test generation, execution, and debugging capabilities for web applications. It generates test cases using AI analysis of application code and UI, executes tests in containerized environments with Selenium, captures detailed execution traces and screenshots, and provides intelligent debugging suggestions when tests fail. The service integrates with Kubernetes for dynamic test environment provisioning and S3 for artifact storage.

## Development Commands

### From This Directory
```bash
# Node service commands
pnpm install          # Install dependencies
pnpm dev              # Hot-reload with tsx watch
pnpm test             # Run tests with Vitest
pnpm test:watch       # Watch mode for tests
pnpm test:ui          # Vitest UI for interactive testing
pnpm test:coverage    # Generate coverage report
pnpm typecheck        # Type checking without building
pnpm lint             # Lint code (not yet configured)
pnpm build            # Build for production
pnpm start            # Run production build
```

### From Root (using Turbo)
```bash
bunx turbo run dev --filter=test-lab-service
bunx turbo run test --filter=test-lab-service
bunx turbo run build --filter=test-lab-service
```

### Pre-PR Checklist
```bash
# Run all quality gates
pnpm typecheck && pnpm lint && pnpm test && pnpm build

# Or from root with turbo
bunx turbo run typecheck lint test build --filter=test-lab-service
```

## Architecture

### Directory Structure
```
test-lab-service/
├── src/
│   ├── __tests__/               # Test suites
│   │   ├── config.test.ts
│   │   ├── test-runs.test.ts
│   │   └── scaffolds.test.ts
│   ├── clients/                 # External service clients
│   │   ├── k8s.client.ts       # Kubernetes client for test env provisioning
│   │   └── s3.client.ts        # S3 client for artifact storage
│   ├── events/                  # NATS event handlers
│   │   ├── publisher.ts
│   │   └── subscriber.ts
│   ├── lib/                     # Utilities
│   │   ├── logger.ts
│   │   └── telemetry.ts
│   ├── middleware/              # Fastify middleware
│   │   ├── auth.middleware.ts
│   │   └── error.middleware.ts
│   ├── repo/                    # Database repositories
│   │   ├── test-runs.repo.ts
│   │   ├── scaffolds.repo.ts
│   │   └── previews.repo.ts
│   ├── routes/                  # REST endpoint handlers
│   │   ├── test-runs.routes.ts
│   │   ├── scaffolds.routes.ts
│   │   └── previews.routes.ts
│   ├── schemas/                 # Zod validation schemas
│   │   └── test-run.schema.ts
│   ├── services/                # Business logic
│   │   ├── test-generator.service.ts
│   │   ├── test-executor.service.ts
│   │   └── debugger.service.ts
│   ├── types/                   # TypeScript types
│   │   └── index.ts
│   ├── app.ts                   # Fastify app setup
│   ├── config.ts                # Configuration management
│   └── index.ts                 # Entry point
├── dist/                        # Build output
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── vitest.config.ts             # Test configuration
├── Makefile                     # Build shortcuts
├── CONTEXT.md                   # Service metadata
└── AGENTS.md                    # AI development shortcuts
```

### Key Components

**Test Generation:**
- File: `src/services/test-generator.service.ts` - AI-powered test case generation
- Pattern: Analyzes application code/UI and generates Selenium test scripts
- Example: Given a URL, generates tests for critical user flows (login, checkout, etc.)

**Test Execution:**
- File: `src/services/test-executor.service.ts` - Orchestrates test runs
- Pattern: Provisions Kubernetes pods with Selenium, executes tests, captures artifacts
- Example: Spins up isolated Chrome browser in K8s, runs test, stores screenshots to S3

**Debugging Service:**
- File: `src/services/debugger.service.ts` - Analyzes test failures
- Pattern: Uses AI to analyze failure traces and suggest fixes
- Example: Detects flaky selectors, timing issues, and recommends remediation

**Kubernetes Client:**
- File: `src/clients/k8s.client.ts` - Manages test environment pods
- Pattern: Creates/deletes Selenium pods on-demand for isolated test execution
- Example: Uses `@kubernetes/client-node` to provision Chrome+Selenium containers

**S3 Client:**
- File: `src/clients/s3.client.ts` - Stores test artifacts
- Pattern: Uploads screenshots, videos, logs to S3 with expiration policies
- Example: Stores failure screenshots at `s3://test-artifacts/{run-id}/screenshots/`

### Dependencies

**Core:**
- `fastify` 4.27.0 - Fast web framework for REST API
- `@fastify/helmet` ^11.1.1 - Security headers middleware
- `@fastify/cors` ^8.4.2 - CORS support
- `@fastify/jwt` ^8.0.0 - JWT authentication
- `pg` ^8.11.3 - PostgreSQL client for test run metadata
- `redis` ^4.6.12 - Redis for caching and rate limiting
- `nats` ^2.25.0 - NATS messaging for event-driven workflows
- `zod` ^3.22.4 - Schema validation for requests

**External Integrations:**
- `@aws-sdk/client-s3` ^3.676.0 - AWS S3 for test artifact storage
- `@kubernetes/client-node` ^1.0.0 - Kubernetes API client for pod provisioning
- `selenium-webdriver` ^4.25.0 - Browser automation for test execution

**Observability:**
- `@opentelemetry/api` ^1.8.0 - OpenTelemetry tracing
- `@opentelemetry/sdk-node` ^0.49.1 - OTel SDK
- `@opentelemetry/auto-instrumentations-node` ^0.42.0 - Auto instrumentation
- `@opentelemetry/instrumentation-fastify` ^0.35.0 - Fastify tracing
- `@opentelemetry/instrumentation-pg` ^0.40.0 - PostgreSQL tracing
- `prom-client` ^15.1.0 - Prometheus metrics
- `pino` ^8.17.2 - Structured logging
- `pino-pretty` ^10.3.1 - Pretty logs for development

**Testing:**
- `vitest` ^1.1.0 - Fast unit test framework
- `@vitest/ui` ^1.1.0 - Interactive test UI
- `@vitest/coverage-v8` ^1.1.0 - Code coverage
- `supertest` ^6.3.3 - HTTP assertions
- `pg-mem` ^3.0.5 - In-memory PostgreSQL for tests
- `jsonwebtoken` ^9.0.2 - JWT mocking for auth tests

## Code Organization Patterns

### Request Handlers
✅ **DO:** Use route -> service -> repository pattern
```typescript
// src/routes/test-runs.routes.ts
export async function testRunsRoutes(fastify: FastifyInstance) {
  fastify.post('/test-runs', async (request, reply) => {
    const body = TestRunSchema.parse(request.body);
    const result = await testExecutorService.createRun(body);
    return reply.send(result);
  });
}
```
❌ **DON'T:** Put business logic in route handlers

### Database Access
✅ **DO:** Use repository pattern with async/await
```typescript
// src/repo/test-runs.repo.ts
export async function createTestRun(run: TestRun): Promise<string> {
  const result = await db.query(
    'INSERT INTO test_runs (project_id, status) VALUES ($1, $2) RETURNING id',
    [run.projectId, 'pending']
  );
  return result.rows[0].id;
}
```
❌ **DON'T:** Write SQL in service layer; always use repositories

### Error Handling
✅ **DO:** Use typed errors and centralized error middleware
```typescript
// src/middleware/error.middleware.ts
fastify.setErrorHandler((error, request, reply) => {
  logger.error({ err: error, traceId: request.id }, 'Request error');
  return reply.status(error.statusCode || 500).send({
    error: { code: error.code, message: error.message, traceId: request.id }
  });
});
```
❌ **DON'T:** Swallow errors or return inconsistent error formats

### Schema Validation
✅ **DO:** Use Zod schemas for runtime validation
```typescript
// src/schemas/test-run.schema.ts
import { z } from 'zod';
export const TestRunSchema = z.object({
  projectId: z.string().uuid(),
  url: z.string().url(),
  testType: z.enum(['e2e', 'visual', 'accessibility']),
});
```
❌ **DON'T:** Skip validation or use manual type checks

## API Endpoints

### REST API

**Base URL:** `http://localhost:7202`

**Key Endpoints:**
- `GET /healthz` - Health check with dependency status
- `GET /metrics` - Prometheus metrics
- `POST /api/v1/test-runs` - Create and execute new test run
- `GET /api/v1/test-runs/{id}` - Get test run status and results
- `GET /api/v1/test-runs/{id}/artifacts` - Download test artifacts (screenshots, logs)
- `POST /api/v1/scaffolds/generate` - Generate test scaffold using AI
- `GET /api/v1/scaffolds/{id}` - Get generated scaffold code
- `POST /api/v1/previews/debug` - Analyze failed test and get debugging suggestions
- `GET /api/v1/previews/{id}` - Get preview environment metadata

### gRPC API

**Port:** 50072

**Services:**
- `TestLabService` - Test execution and debugging
  - `CreateTestRun` - Execute tests in isolated environment
  - `GetTestResults` - Retrieve test run results
  - `GenerateTests` - AI-powered test generation
  - `DebugFailure` - Analyze test failures

**Proto Files:** `api/*.proto` (if present)

## Database Schema

**Tables:**

**`test_runs`** - Test execution metadata
- Columns: `id`, `project_id`, `url`, `test_type`, `status`, `started_at`, `completed_at`, `created_by`
- Indexes: `idx_project_id`, `idx_status`, `idx_created_at`
- Purpose: Track all test executions and their status

**`test_results`** - Individual test case results
- Columns: `id`, `test_run_id`, `test_name`, `status`, `duration_ms`, `error_message`, `screenshot_url`, `video_url`
- Indexes: `idx_test_run_id`, `idx_status`
- Purpose: Detailed results for each test case within a run

**`scaffolds`** - Generated test scaffolds
- Columns: `id`, `project_id`, `scaffold_type`, `code`, `language`, `framework`, `created_at`
- Indexes: `idx_project_id`, `idx_scaffold_type`
- Purpose: Store AI-generated test code for reuse

**`previews`** - Preview environment metadata
- Columns: `id`, `test_run_id`, `k8s_pod_name`, `url`, `status`, `created_at`, `expires_at`
- Indexes: `idx_test_run_id`, `idx_k8s_pod_name`
- Purpose: Track ephemeral test environments in Kubernetes

**Migrations:**
- Location: `src/db/migrations/` (if present)
- Tool: Custom migration scripts
- Commands: `pnpm db:migrate:up`, `pnpm db:migrate:down`

## Event Handling

**Published Events:**
- `test-run.created` - When new test run starts
  - Payload: `{test_run_id, project_id, url, test_type}`
- `test-run.completed` - When test run finishes
  - Payload: `{test_run_id, status, total_tests, passed, failed, duration_ms}`
- `test-run.failed` - When test run encounters critical error
  - Payload: `{test_run_id, error_message, stack_trace}`
- `scaffold.generated` - When AI generates new test scaffold
  - Payload: `{scaffold_id, project_id, framework, lines_of_code}`

**Subscribed Events:**
- `project.created` - Initialize test configuration for new project
- `deployment.completed` - Trigger smoke tests for newly deployed environments

## Testing Strategy

### Unit Tests
- Location: `src/__tests__/*.test.ts` - Colocated with source modules
- Coverage: Target >80%
- Mock: Database with `pg-mem`, external services with Vitest mocks
- Example: `src/__tests__/test-runs.test.ts`

### Integration Tests
- Location: `src/__tests__/integration/` (if present)
- Setup: Use Testcontainers for PostgreSQL, Redis
- Pattern: Test full request flow from route -> service -> database

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
```

## Configuration

### Environment Variables
```bash
# Service-specific config
K8S_NAMESPACE=test-lab
K8S_SELENIUM_IMAGE=selenium/standalone-chrome:latest
TEST_TIMEOUT_SECONDS=300
MAX_CONCURRENT_TESTS=10
S3_BUCKET=test-lab-artifacts
S3_ARTIFACT_EXPIRY_DAYS=30
AI_MODEL_ENDPOINT=http://ai-debugger-service:7102

# Standard config (inherited from ../CLAUDE.md)
SERVICE_NAME=test-lab-service
SERVICE_PORT=7202
GRPC_PORT=50072
ENV=dev
DATABASE_URL=postgresql://user:pass@localhost:5432/testlab
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
VAULT_ADDR=http://vault:8200
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
LOG_LEVEL=info
```

### Secrets
- Stored in: Vault at `secret/test-lab-service/`
- Accessed via: `secrets-env-service` or direct Vault SDK
- Keys: AWS credentials for S3, Kubernetes service account tokens

## Quick Find Commands

### Find Code
```bash
# Find test execution logic
rg -n "executeTest|runTest" services/test-lab-service/src/

# Find Kubernetes pod creation
rg -n "createPod|Pod.*create" services/test-lab-service/src/clients/

# Find S3 artifact uploads
rg -n "s3.*upload|putObject" services/test-lab-service/src/clients/

# Find event publishers
rg -n "publish.*test-run\." services/test-lab-service/src/events/

# Find Selenium usage
rg -n "webdriver|selenium" services/test-lab-service/src/
```

### Find Dependencies
```bash
# Check what depends on this service
rg -n "test-lab-service" --glob "docker-compose*.yml" --glob "*.yaml"

# Find route definitions
rg -n "fastify\.(get|post)" services/test-lab-service/src/routes/
```

## Common Gotchas

- **Kubernetes Pod Cleanup:** Failed tests may leave orphaned pods; ensure cleanup logic runs in finally blocks or use pod TTL with `ttlSecondsAfterFinished`
- **Selenium Timeouts:** Slow pages may exceed default timeouts; adjust `TEST_TIMEOUT_SECONDS` or implement smart waits in test scripts
- **S3 Artifact Size:** Large videos can consume storage quickly; configure lifecycle policies to delete old artifacts after `S3_ARTIFACT_EXPIRY_DAYS`
- **Concurrent Test Limits:** Running too many Selenium instances can exhaust cluster resources; tune `MAX_CONCURRENT_TESTS` based on K8s capacity
- **Flaky Tests:** Network issues or timing problems cause flakiness; implement retry logic and better wait conditions in generated tests
- **Secret Scanning:** Avoid hardcoded credentials in test code; use placeholders and inject from Vault at runtime

## Related Services

- **ai-debugger-service:** Provides AI model for analyzing test failures and generating debugging suggestions
- **projects-service:** Manages project metadata; Test Lab queries project configuration for test settings
- **notification-service:** Sends alerts when test runs fail or when critical issues detected
- **observability-service:** Aggregates test run metrics and failure trends for dashboards

## Useful Links

- Service Catalog: `../../service_catalog.yaml`
- CONTEXT.md: Service purpose, SLOs, ownership details
- AGENTS.md: AI-assisted development shortcuts
- CI Workflow: `../../.github/workflows/ci-test-lab-service.yml`
