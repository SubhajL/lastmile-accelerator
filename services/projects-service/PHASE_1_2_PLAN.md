# projects-service: Phase 1 & 2 Implementation Plan

## Overview

Build the foundation (Phase 1) and core APIs (Phase 2) for projects-service using **TDD first** principles. Phase 1 establishes the Fastify server, database layer, and observability scaffolding. Phase 2 implements all CRUD endpoints for projects, tenants, members, and environments with full tenant isolation and JWT auth.

**Architecture:** Node.js + Fastify → PostgreSQL (via pg library) → NATS (via nats.js) → OTel (via @opentelemetry/*)

**Test Framework:** Vitest (fast, ESM-native, TypeScript support)

**Secrets:** Environment variables + Vault integration (env vars only in Phase 1)

---

## Files to Create/Modify

### **New Files**

```
src/
├── config.ts                      # Environment loading & validation
├── logger.ts                      # Pino logger setup
├── db.ts                          # PostgreSQL connection pool
├── nats.ts                        # NATS publisher wrapper
├── otel.ts                        # OpenTelemetry setup
├── middleware/
│   ├── auth.ts                    # JWT verification middleware
│   ├── errorHandler.ts            # Global error handler
│   ├── requestId.ts               # Request tracing middleware
│   └── metrics.ts                 # Prometheus middleware
├── routes/
│   ├── health.ts                  # /healthz + /metrics endpoints
│   ├── projects.ts                # Project CRUD routes
│   ├── tenants.ts                 # Tenant routes
│   ├── members.ts                 # Member management routes
│   └── environments.ts            # Environment config routes
├── db/
│   ├── schema.ts                  # TypeScript types matching DB schema
│   ├── migrations/
│   │   ├── 001_init.sql           # Create all tables
│   │   └── migrate.ts             # Migration runner
│   └── seed.ts                    # (Optional) Test data seeder
├── services/
│   ├── projectService.ts          # Project business logic
│   ├── tenantService.ts           # Tenant business logic
│   ├── memberService.ts           # Member business logic
│   └── environmentService.ts      # Environment business logic
├── events/
│   └── eventPublisher.ts          # Publish NATS events
├── utils/
│   ├── errors.ts                  # Custom error classes
│   ├── validators.ts              # Input validation (zod schemas)
│   └── helpers.ts                 # Utility functions
└── __tests__/
    ├── setup.ts                   # Test configuration & fixtures
    ├── fixtures/
    │   ├── tenants.fixture.ts     # Mock tenant data
    │   ├── projects.fixture.ts    # Mock project data
    │   └── users.fixture.ts       # Mock user data
    ├── unit/
    │   ├── services/
    │   │   ├── projectService.test.ts
    │   │   ├── tenantService.test.ts
    │   │   ├── memberService.test.ts
    │   │   └── environmentService.test.ts
    │   ├── middleware/
    │   │   └── auth.test.ts
    │   └── utils/
    │       └── validators.test.ts
    └── integration/
        ├── projects.routes.test.ts
        ├── tenants.routes.test.ts
        ├── members.routes.test.ts
        └── environments.routes.test.ts

index.ts (existing - rewrite with Fastify + middleware setup)
```

### **Modified Files**

```
package.json                        # Add dependencies + test script
tsconfig.json                       # Add test include + stricter rules
```

---

## Phase 1: Foundation

### 1.1 Dependencies Setup

**Update `package.json`:**

```json
{
  "dependencies": {
    "fastify": "4.27.0",
    "@fastify/helmet": "^11.1.1",
    "@fastify/cors": "^8.4.2",
    "pg": "^8.11.3",
    "nats": "^2.25.0",
    "@opentelemetry/api": "^1.7.0",
    "@opentelemetry/sdk-node": "^0.48.0",
    "@opentelemetry/sdk-trace-node": "^0.48.0",
    "@opentelemetry/auto-instrumentations-node": "^0.47.1",
    "@opentelemetry/exporter-trace-otlp-http": "^0.48.0",
    "pino": "^8.17.2",
    "pino-http": "^8.7.0",
    "zod": "^3.22.4",
    "jsonwebtoken": "^9.1.2",
    "uuid": "^9.0.1"
  },
  "devDependencies": {
    "tsx": "4.19.2",
    "typescript": "5.6.3",
    "@types/node": "^20.10.6",
    "@types/pg": "^8.11.2",
    "vitest": "^1.1.0",
    "@vitest/ui": "^1.1.0",
    "supertest": "^6.3.3",
    "@types/supertest": "^6.0.2",
    "@testing-library/node": "^0.2.3",
    "ts-node": "^10.9.2"
  },
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc -p tsconfig.json",
    "start": "node --enable-source-maps dist/index.js",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage",
    "migrate": "tsx src/db/migrations/migrate.ts",
    "seed": "tsx src/db/seed.ts"
  }
}
```

### 1.2 Configuration Layer

**`src/config.ts`**

```typescript
// loadConfig(): ServerConfig
// Reads SERVICE_PORT, SERVICE_NAME, DATABASE_URL, NATS_URL, OTEL_EXPORTER_OTLP_ENDPOINT
// Returns validated config object; throws on missing required vars.

// validateEnv(): void
// Uses zod to validate all required env vars at startup; fails fast.
```

**Test:** `config.test.ts` → test missing env var throws, valid config passes.

### 1.3 Logger Setup

**`src/logger.ts`**

```typescript
// createLogger(serviceName: string): PinoLogger
// Initializes Pino logger with JSON formatting, request ID tracking, structured logging.

// getLogger(): PinoLogger
// Returns singleton logger instance.
```

**Test:** `logger.test.ts` → logs are JSON, include service.name field.

### 1.4 Database Layer

**`src/db.ts`**

```typescript
// initDb(config: DbConfig): Promise<Pool>
// Creates PostgreSQL connection pool with connection string validation; retries on failure.

// getDb(): Pool
// Returns singleton connection pool.

// closeDb(): Promise<void>
// Gracefully closes all connections.

// query(sql: string, params?: any[]): Promise<QueryResult>
// Executes query with automatic pool management and error handling.

// transaction<T>(callback: (client: PoolClient) => Promise<T>): Promise<T>
// Wraps callback in BEGIN/COMMIT/ROLLBACK; atomicity guarantee.
```

**`src/db/migrations/001_init.sql`**

```sql
-- Create tenants table
CREATE TABLE tenants (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP NULL
);

-- Create projects table
CREATE TABLE projects (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP NULL
);

-- Create members table (tenant membership with roles)
CREATE TABLE members (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  user_id UUID NOT NULL,
  email VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL, -- owner, admin, developer, viewer
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Create environments table
CREATE TABLE environments (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  name VARCHAR(50) NOT NULL, -- dev, staging, prod
  config JSONB DEFAULT '{}',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Create ingestion_modes table
CREATE TABLE ingestion_modes (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  mode VARCHAR(10) NOT NULL, -- A, B, or C
  is_default BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for tenant isolation
CREATE INDEX idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX idx_members_tenant_id ON members(tenant_id);
CREATE INDEX idx_environments_project_id ON environments(project_id);
CREATE INDEX idx_ingestion_modes_project_id ON ingestion_modes(project_id);
```

**`src/db/migrations/migrate.ts`**

```typescript
// runMigrations(): Promise<void>
// Reads .sql files in migrations/ directory, executes in order, tracks applied migrations.
```

**Tests:** `db.test.ts` → DB pool connects, query executes, transaction rolls back on error.

### 1.5 NATS Publisher

**`src/nats.ts`**

```typescript
// initNats(url: string): Promise<NatsConnection>
// Connects to NATS server; retries with exponential backoff.

// getNats(): NatsConnection
// Returns singleton connection.

// publish(topic: string, payload: Record<string, any>, traceparent?: string): Promise<void>
// Publishes message with traceparent header for distributed tracing.

// closeNats(): Promise<void>
// Gracefully closes NATS connection.
```

**Test:** `nats.test.ts` → connects to NATS, publishes message, adds traceparent header.

### 1.6 OpenTelemetry Setup

**`src/otel.ts`**

```typescript
// initOtel(serviceName: string): Promise<void>
// Initializes OTel SDK with OTLP exporter; sets global tracer & meter.

// getTracer(name?: string): Tracer
// Returns OTel tracer for manual span creation.

// getMeter(name?: string): Meter
// Returns OTel meter for custom metrics.
```

**Test:** `otel.test.ts` → tracer initialized, spans exported, service.name tagged.

### 1.7 Authentication Middleware

**`src/middleware/auth.ts`**

```typescript
// verifyJwt(token: string, config: AuthConfig): Promise<JwtPayload>
// Validates JWT signature using JWKS URL; throws on invalid/expired token.

// authMiddleware(): FastifyReply
// Fastify middleware that extracts Authorization header, verifies JWT, attaches user to request.
// Scopes validation per route (done in route handlers).

// extractTenantId(request: FastifyRequest): string
// Extracts tenant_id from JWT payload; throws 401 if missing.
```

**Test:** `auth.test.ts` → valid JWT accepted, invalid rejected, tenant_id extracted.

### 1.8 Health & Metrics Endpoints

**`src/routes/health.ts`**

```typescript
// registerHealthRoutes(app: FastifyInstance): void
// Registers:
// - GET /healthz → {status: "ok", timestamp}
// - GET /metrics → Prometheus format with service.name tag

// Metrics to track: request_count, request_duration_ms, error_rate
```

**Test:** `health.routes.test.ts` → /healthz responds 200, /metrics has Prometheus headers.

### 1.9 Global Error Handler

**`src/middleware/errorHandler.ts`**

```typescript
// errorHandler(error: FastifyError, request: FastifyRequest, reply: FastifyReply): void
// Catches all errors, formats as {code, message, details}, logs with request ID, returns appropriate HTTP status.

// Custom error classes: ValidationError(400), AuthError(401), NotFoundError(404), ConflictError(409)
```

**Test:** `errorHandler.test.ts` → 400 for validation, 401 for auth, 404 for not found.

### 1.10 Request ID & Tracing Middleware

**`src/middleware/requestId.ts`**

```typescript
// requestIdMiddleware(): void
// Generates or extracts request ID from header; adds to request context & response headers.
// Used for end-to-end tracing across services.
```

**Test:** `requestId.test.ts` → request ID added to header, propagated to logs.

### 1.11 Main Server Bootstrap

**`src/index.ts` (rewrite)**

```typescript
// Main entry point that:
// 1. Loads config via loadConfig()
// 2. Initializes logger
// 3. Initializes DB + runs migrations
// 4. Initializes NATS
// 5. Initializes OTel
// 6. Creates Fastify instance with middleware (helmet, cors, auth, errorHandler, requestId)
// 7. Registers routes (health, projects, tenants, members, environments)
// 8. Listens on SERVICE_PORT
// 9. Handles graceful shutdown (SIGTERM/SIGINT)
```

---

## Phase 2: Core APIs

### 2.1 Database Schema Types

**`src/db/schema.ts`**

```typescript
// Define TypeScript interfaces matching DB schema:
// - Tenant {id, name, slug, createdAt, updatedAt, deletedAt}
// - Project {id, tenantId, name, description, createdAt, updatedAt, deletedAt}
// - Member {id, tenantId, userId, email, role, createdAt, updatedAt}
// - Environment {id, projectId, name, config, createdAt, updatedAt}
// - IngestionMode {id, projectId, mode, isDefault, createdAt}
```

### 2.2 Validation Schemas

**`src/utils/validators.ts`**

```typescript
// Zod schemas for:
// - CreateProjectInput {name, description?}
// - UpdateProjectInput {name?, description?}
// - CreateMemberInput {email, role}
// - CreateEnvironmentInput {name, config?}
// - SetIngestionModeInput {mode: "A"|"B"|"C", isDefault?}

// Validates shape, types, required fields; throws on invalid input.
```

**Test:** `validators.test.ts` → valid input passes, invalid input throws ZodError.

### 2.3 Project Service

**`src/services/projectService.ts`**

```typescript
// listProjects(tenantId: string): Promise<Project[]>
// Queries all non-deleted projects for tenant; ordered by createdAt DESC.

// getProject(tenantId: string, projectId: string): Promise<Project>
// Fetches single project; validates tenant_id match; throws NotFoundError if missing.

// createProject(tenantId: string, input: CreateProjectInput, userId: string): Promise<Project>
// Inserts project, creates default environments (dev, staging, prod), adds creator as owner member.
// Wrapped in transaction; publishes project.created event.

// updateProject(tenantId: string, projectId: string, input: UpdateProjectInput): Promise<Project>
// Updates name/description; validates tenant match; publishes project.updated event.

// deleteProject(tenantId: string, projectId: string): Promise<void>
// Soft-deletes (sets deleted_at); publishes project.deleted event.
```

**Tests:** `projectService.test.ts`
- `createProject creates project with default environments`
- `getProject returns project for correct tenant`
- `getProject throws 404 for wrong tenant`
- `listProjects returns only non-deleted projects for tenant`
- `updateProject publishes project.updated event`
- `deleteProject soft-deletes and publishes event`

### 2.4 Tenant Service

**`src/services/tenantService.ts`**

```typescript
// getTenant(tenantId: string): Promise<Tenant>
// Fetches tenant; throws NotFoundError if not found or deleted.

// listTenantMembers(tenantId: string): Promise<Member[]>
// Queries all members for tenant; ordered by email.

// Can be extended later for tenant creation (admin-only in v1).
```

**Tests:** `tenantService.test.ts`
- `getTenant returns tenant`
- `getTenant throws 404 for deleted tenant`
- `listTenantMembers returns all members`

### 2.5 Member Service

**`src/services/memberService.ts`**

```typescript
// addMember(tenantId: string, input: CreateMemberInput, requesterRole: string): Promise<Member>
// Validates requester is owner; generates new user_id UUID; inserts member.
// Publishes member.added event.

// getMember(tenantId: string, memberId: string): Promise<Member>
// Fetches member; validates tenant_id match; throws NotFoundError if missing.

// updateMemberRole(tenantId: string, memberId: string, newRole: string, requesterRole: string): Promise<Member>
// Validates requester is owner; updates role; publishes member.role_changed event.

// removeMember(tenantId: string, memberId: string, requesterRole: string): Promise<void>
// Validates requester is owner; prevents removing last owner; soft-deletes member.
// Publishes member.removed event.
```

**Tests:** `memberService.test.ts`
- `addMember creates member and publishes event`
- `addMember throws error if requester not owner`
- `updateMemberRole updates and publishes event`
- `removeMember prevents removing last owner`
- `getMember throws 404 for wrong tenant`

### 2.6 Environment Service

**`src/services/environmentService.ts`**

```typescript
// listEnvironments(tenantId: string, projectId: string): Promise<Environment[]>
// Queries all environments for project; validates project tenant match; ordered by name.

// getEnvironment(tenantId: string, projectId: string, envId: string): Promise<Environment>
// Fetches environment; validates tenant + project match; throws NotFoundError if missing.

// createEnvironment(tenantId: string, projectId: string, input: CreateEnvironmentInput): Promise<Environment>
// Inserts environment; validates project tenant match; publishes environment.created event.

// updateEnvironment(tenantId: string, projectId: string, envId: string, input: UpdateEnvironmentInput): Promise<Environment>
// Updates config JSONB; validates tenant + project match; publishes environment.updated event.

// deleteEnvironment(tenantId: string, projectId: string, envId: string): Promise<void>
// Deletes environment; prevents deleting if < 2 remain; publishes environment.deleted event.

// setIngestionModes(tenantId: string, projectId: string, modes: string[], defaultMode: string): Promise<void>
// Validates modes are from {A, B, C}; sets default; publishes ingestion_modes.updated event.

// getIngestionModes(tenantId: string, projectId: string): Promise<IngestionMode[]>
// Returns modes for project; sorted with isDefault first.
```

**Tests:** `environmentService.test.ts`
- `createEnvironment creates env and publishes event`
- `listEnvironments returns only envs for project tenant`
- `setIngestionModes updates modes and publishes event`
- `deleteEnvironment prevents deleting last env`
- `updateEnvironment validates tenant + project match`

### 2.7 Event Publisher

**`src/events/eventPublisher.ts`**

```typescript
// publishProjectEvent(event: ProjectEvent, traceparent?: string): Promise<void>
// Publishes to NATS topic: project.created | project.updated | project.deleted
// Payload: {projectId, tenantId, ...eventData, timestamp, traceparent}

// publishMemberEvent(event: MemberEvent, traceparent?: string): Promise<void>
// Publishes to NATS topic: member.added | member.updated | member.removed
// Payload: {memberId, tenantId, ...eventData, timestamp, traceparent}

// publishEnvironmentEvent(event: EnvironmentEvent, traceparent?: string): Promise<void>
// Publishes to NATS topic: environment.created | environment.updated | environment.deleted
// Payload: {environmentId, projectId, tenantId, ...eventData, timestamp, traceparent}
```

**Tests:** `eventPublisher.test.ts`
- `publishProjectEvent publishes to correct NATS topic`
- `Events include traceparent for tracing`
- `Handles NATS publish errors gracefully`

### 2.8 Route Handlers: Projects

**`src/routes/projects.ts`**

```typescript
// registerProjectRoutes(app: FastifyInstance): void
// Routes:
// - GET /v1/projects → List projects for tenant
// - POST /v1/projects → Create project
// - GET /v1/projects/{projectId} → Get project
// - PUT /v1/projects/{projectId} → Update project
// - DELETE /v1/projects/{projectId} → Delete project

// Each route:
// 1. Verifies auth via authMiddleware (extracts tenantId)
// 2. Validates input (if POST/PUT)
// 3. Calls service layer
// 4. Returns JSON response
// 5. Errors caught by errorHandler
```

**Tests:** `projects.routes.test.ts`
- `GET /v1/projects returns 200 with list`
- `POST /v1/projects creates and returns 201`
- `GET /v1/projects/{id} returns 200`
- `PUT /v1/projects/{id} updates and returns 200`
- `DELETE /v1/projects/{id} returns 204`
- `GET /v1/projects/{id} from other tenant returns 403`
- `Invalid input returns 400`

### 2.9 Route Handlers: Tenants

**`src/routes/tenants.ts`**

```typescript
// registerTenantRoutes(app: FastifyInstance): void
// Routes:
// - GET /v1/tenants/{tenantId} → Get tenant
// - GET /v1/tenants/{tenantId}/members → List members
```

**Tests:** `tenants.routes.test.ts`
- `GET /v1/tenants/{id} returns 200`
- `GET /v1/tenants/{id} for other tenant returns 403`
- `GET /v1/tenants/{id}/members returns list of members`

### 2.10 Route Handlers: Members

**`src/routes/members.ts`**

```typescript
// registerMemberRoutes(app: FastifyInstance): void
// Routes:
// - POST /v1/tenants/{tenantId}/members → Add member
// - PUT /v1/members/{memberId} → Update role
// - DELETE /v1/members/{memberId} → Remove member
```

**Tests:** `members.routes.test.ts`
- `POST /v1/tenants/{id}/members creates and returns 201`
- `POST with non-owner user returns 403`
- `PUT /v1/members/{id} updates role and returns 200`
- `DELETE /v1/members/{id} returns 204`
- `DELETE last owner returns 409 (conflict)`

### 2.11 Route Handlers: Environments

**`src/routes/environments.ts`**

```typescript
// registerEnvironmentRoutes(app: FastifyInstance): void
// Routes:
// - GET /v1/projects/{projectId}/environments → List environments
// - POST /v1/projects/{projectId}/environments → Create environment
// - GET /v1/projects/{projectId}/environments/{envId} → Get environment
// - PUT /v1/projects/{projectId}/environments/{envId} → Update environment
// - DELETE /v1/projects/{projectId}/environments/{envId} → Delete environment
// - GET /v1/projects/{projectId}/ingestion-modes → Get ingestion modes
// - POST /v1/projects/{projectId}/ingestion-modes → Set modes
```

**Tests:** `environments.routes.test.ts`
- `GET /v1/projects/{id}/environments returns 200 with list`
- `POST /v1/projects/{id}/environments creates and returns 201`
- `PUT /v1/projects/{id}/environments/{envId} updates and returns 200`
- `DELETE /v1/projects/{id}/environments/{envId} returns 204`
- `DELETE last environment returns 409`
- `POST ingestion-modes sets modes and returns 200`

### 2.12 Test Setup & Fixtures

**`src/__tests__/setup.ts`**

```typescript
// globalSetup(): void
// Before all tests: connect to test DB, run migrations, seed fixtures.

// globalTeardown(): void
// After all tests: rollback, close DB connection.

// withDbTransaction(test: () => Promise<void>): Promise<void>
// Wraps test in transaction that rolls back after test (isolation).
```

**`src/__tests__/fixtures/tenants.fixture.ts`**

```typescript
// createTestTenant(overrides?: Partial<Tenant>): Promise<Tenant>
// Inserts test tenant; returns tenant record.

// createTestMember(tenantId: string, overrides?: Partial<Member>): Promise<Member>
// Inserts test member for tenant.
```

Similarly for `projects.fixture.ts`, `users.fixture.ts`.

---

## Implementation Order (Recommended)

1. **Setup**: package.json, config.ts, logger.ts
2. **Database**: db.ts, migrations, schema.ts
3. **Observability**: otel.ts, nats.ts
4. **Middleware**: auth.ts, errorHandler.ts, requestId.ts
5. **Routes Foundation**: health.ts (verify all plumbing works)
6. **Services**: projectService → tenantService → memberService → environmentService (unit tests first)
7. **Route Handlers**: projects.ts → tenants.ts → members.ts → environments.ts (integration tests)
8. **Polish**: test coverage, docs, graceful shutdown

---

## Key Design Decisions

- **Tenant Isolation**: All queries filter by tenant_id at service layer; enforced via foreign keys.
- **TDD First**: Write tests before implementing; green tests gate each feature.
- **Event-Driven**: Every mutation publishes NATS event for downstream listeners.
- **Transaction Safety**: Multi-step operations wrapped in transactions.
- **Error Handling**: Custom error classes map to HTTP status; caught by global handler.
- **Type Safety**: Strict TypeScript; Zod validation for inputs.
- **Observability**: OTel spans + Pino logs + Prometheus metrics on every endpoint.

---

## Success Criteria

- ✅ All Phase 1 tests pass (db, logger, otel, middleware)
- ✅ All Phase 2 tests pass (services + routes)
- ✅ Service starts locally: `SERVICE_PORT=7002 pnpm dev`
- ✅ Health check responds: `curl http://localhost:7002/healthz`
- ✅ Metrics exposed: `curl http://localhost:7002/metrics`
- ✅ Test coverage > 80%
- ✅ TypeScript compilation clean: `npm run build`
