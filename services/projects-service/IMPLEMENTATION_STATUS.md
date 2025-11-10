# projects-service: Implementation Status

## ✅ Phase 1: Completed

### Foundation Layer (TDD Complete)

#### 1. Configuration (✅ Done)
- **File:** `src/config.ts`
- **Tests:** `src/__tests__/unit/config.test.ts` (7 tests passing)
- **Status:** ✅ All tests passing, TypeScript compiles
- **Functions:**
  - `loadConfig()` — Loads and validates env vars via Zod
  - `validateEnv()` — Fails fast on missing vars
  - `getConfig()` — Singleton accessor

#### 2. Logging (✅ Done)
- **File:** `src/logger.ts`
- **Tests:** `src/__tests__/unit/logger.test.ts` (7 tests passing)
- **Status:** ✅ All tests passing, TypeScript compiles
- **Functions:**
  - `createLogger(serviceName)` — Initialize Pino logger with structured JSON
  - `getLogger()` — Singleton accessor
  - `getContextLogger(fields)` — Child logger for request context

### Remaining Phase 1 Files (To Complete)

#### 3. Database Layer
- **File:** `src/db.ts`
- **Tests:** `src/__tests__/unit/db.test.ts`
- **Functions to implement:**
  - `initDb(config)` — Create PostgreSQL pool with retries
  - `getDb()` — Singleton accessor
  - `query(sql, params)` — Execute query
  - `transaction<T>(callback)` — Atomic transaction wrapper

#### 4. Database Migrations
- **File:** `src/db/migrations/001_init.sql`
- **Runner:** `src/db/migrations/migrate.ts`
- **Purpose:** Create 6 tables (tenants, projects, members, environments, ingestion_modes)

#### 5. NATS Event Publisher
- **File:** `src/nats.ts`
- **Tests:** `src/__tests__/unit/nats.test.ts`
- **Functions:**
  - `initNats(url)` — Connect to NATS JetStream
  - `getNats()` — Singleton accessor
  - `publish(topic, payload, traceparent)` — Publish event with tracing

#### 6. OpenTelemetry Setup
- **File:** `src/otel.ts`
- **Tests:** `src/__tests__/unit/otel.test.ts`
- **Functions:**
  - `initOtel(serviceName)` — Initialize OTel SDK + OTLP exporter
  - `getTracer(name?)` — Get tracer instance
  - `getMeter(name?)` — Get meter instance

#### 7. Middleware Layer
- **Auth:** `src/middleware/auth.ts` + test
  - `verifyJwt(token, config)` — Validate JWT via JWKS
  - `authMiddleware()` — Fastify middleware
  - `extractTenantId(request)` — Extract tenant from JWT

- **Error Handler:** `src/middleware/errorHandler.ts` + test
  - `errorHandler(error, request, reply)` — Global error handler
  - Custom errors: ValidationError, AuthError, NotFoundError, ConflictError

- **Request ID:** `src/middleware/requestId.ts` + test
  - `requestIdMiddleware()` — Generate/extract request ID

- **Metrics:** `src/middleware/metrics.ts` + test
  - Prometheus middleware for request_count, request_duration_ms

#### 8. Error Utilities
- **File:** `src/utils/errors.ts`
- **Classes:**
  - `ValidationError` (400)
  - `AuthError` (401)
  - `NotFoundError` (404)
  - `ConflictError` (409)

#### 9. Health Routes
- **File:** `src/routes/health.ts` + test
- **Endpoints:**
  - `GET /healthz` → {status: "ok", timestamp}
  - `GET /metrics` → Prometheus format

#### 10. Main Server Bootstrap
- **File:** `src/index.ts` (rewrite)
- **Bootstrap sequence:**
  1. Load config
  2. Initialize logger
  3. Initialize DB + run migrations
  4. Initialize NATS
  5. Initialize OTel
  6. Create Fastify instance with middleware
  7. Register routes
  8. Listen on port
  9. Handle graceful shutdown

---

## 📊 Phase 2: Ready to Begin

### Core Database & Services

#### Database Schema (`src/db/schema.ts`)
```typescript
interface Tenant {
  id: string;
  name: string;
  slug: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

interface Project {
  id: string;
  tenantId: string;
  name: string;
  description?: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

interface Member {
  id: string;
  tenantId: string;
  userId: string;
  email: string;
  role: 'owner' | 'admin' | 'developer' | 'viewer';
  createdAt: Date;
  updatedAt: Date;
}

interface Environment {
  id: string;
  projectId: string;
  name: string;
  config: Record<string, any>;
  createdAt: Date;
  updatedAt: Date;
}

interface IngestionMode {
  id: string;
  projectId: string;
  mode: 'A' | 'B' | 'C';
  isDefault: boolean;
  createdAt: Date;
}
```

### Services to Implement (Phase 2)

1. **projectService.ts** + tests
   - `listProjects(tenantId)` → Project[]
   - `getProject(tenantId, projectId)` → Project
   - `createProject(tenantId, input, userId)` → Project (with transaction)
   - `updateProject(tenantId, projectId, input)` → Project
   - `deleteProject(tenantId, projectId)` → void

2. **tenantService.ts** + tests
   - `getTenant(tenantId)` → Tenant
   - `listTenantMembers(tenantId)` → Member[]

3. **memberService.ts** + tests
   - `addMember(tenantId, input, requesterRole)` → Member
   - `getMember(tenantId, memberId)` → Member
   - `updateMemberRole(tenantId, memberId, newRole, requesterRole)` → Member
   - `removeMember(tenantId, memberId, requesterRole)` → void

4. **environmentService.ts** + tests
   - `listEnvironments(tenantId, projectId)` → Environment[]
   - `getEnvironment(tenantId, projectId, envId)` → Environment
   - `createEnvironment(tenantId, projectId, input)` → Environment
   - `updateEnvironment(tenantId, projectId, envId, input)` → Environment
   - `deleteEnvironment(tenantId, projectId, envId)` → void
   - `setIngestionModes(tenantId, projectId, modes, defaultMode)` → void
   - `getIngestionModes(tenantId, projectId)` → IngestionMode[]

### Routes to Implement (Phase 2)

1. **routes/projects.ts** + tests
   - GET /v1/projects
   - POST /v1/projects
   - GET /v1/projects/:projectId
   - PUT /v1/projects/:projectId
   - DELETE /v1/projects/:projectId

2. **routes/tenants.ts** + tests
   - GET /v1/tenants/:tenantId
   - GET /v1/tenants/:tenantId/members

3. **routes/members.ts** + tests
   - POST /v1/tenants/:tenantId/members
   - PUT /v1/members/:memberId
   - DELETE /v1/members/:memberId

4. **routes/environments.ts** + tests
   - GET /v1/projects/:projectId/environments
   - POST /v1/projects/:projectId/environments
   - GET /v1/projects/:projectId/environments/:envId
   - PUT /v1/projects/:projectId/environments/:envId
   - DELETE /v1/projects/:projectId/environments/:envId
   - GET /v1/projects/:projectId/ingestion-modes
   - POST /v1/projects/:projectId/ingestion-modes

### Event Publisher (`src/events/eventPublisher.ts`)
- `publishProjectEvent(event, traceparent)` → void
- `publishMemberEvent(event, traceparent)` → void
- `publishEnvironmentEvent(event, traceparent)` → void

---

## 🎯 Next Steps

### Immediate (Complete Phase 1)

Follow TDD pattern (write test → implement → verify):

1. **Database Setup** (2-3 hours)
   - Create `src/db.ts` with tests
   - Create migration SQL + runner
   - Verify DB pool connects

2. **NATS Setup** (1 hour)
   - Create `src/nats.ts` with tests
   - Verify connection + publish

3. **OTel Setup** (1 hour)
   - Create `src/otel.ts` with tests
   - Verify tracer initialization

4. **Middleware** (2-3 hours)
   - Auth middleware + tests
   - Error handler + tests
   - Request ID middleware + tests
   - Metrics middleware (minimal)

5. **Health Routes & Bootstrap** (1 hour)
   - Create `/healthz` endpoint
   - Create `/metrics` endpoint
   - Wire everything in `index.ts`
   - **CHECKPOINT:** `pnpm dev` starts, `/healthz` responds

### Then (Phase 2)

Once Phase 1 is complete and server boots:

6. **Database Schema & Services** (3-4 hours)
   - Implement all 4 services with unit tests
   - Create fixtures for test data

7. **Route Handlers** (3-4 hours)
   - Implement all 4 route handlers with integration tests
   - Verify auth, tenant isolation, error handling

8. **Polish** (1-2 hours)
   - Test coverage report
   - Documentation
   - Graceful shutdown

---

## 📋 Files Structure

```
src/
├── index.ts                                    # Main server bootstrap
├── config.ts                                   # ✅ DONE
├── logger.ts                                   # ✅ DONE
├── db.ts                                       # TODO
├── nats.ts                                     # TODO
├── otel.ts                                     # TODO
├── middleware/
│   ├── auth.ts                                 # TODO
│   ├── errorHandler.ts                         # TODO
│   ├── requestId.ts                            # TODO
│   ├── metrics.ts                              # TODO (minimal)
│   └── [tests]
├── routes/
│   ├── health.ts                               # TODO
│   ├── projects.ts                             # TODO (Phase 2)
│   ├── tenants.ts                              # TODO (Phase 2)
│   ├── members.ts                              # TODO (Phase 2)
│   ├── environments.ts                         # TODO (Phase 2)
│   └── [tests]
├── db/
│   ├── schema.ts                               # TODO (Phase 2)
│   ├── migrations/
│   │   ├── 001_init.sql                        # TODO
│   │   └── migrate.ts                          # TODO
│   └── seed.ts                                 # TODO (Phase 2)
├── services/
│   ├── projectService.ts                       # TODO (Phase 2)
│   ├── tenantService.ts                        # TODO (Phase 2)
│   ├── memberService.ts                        # TODO (Phase 2)
│   ├── environmentService.ts                   # TODO (Phase 2)
│   └── [tests]
├── events/
│   ├── eventPublisher.ts                       # TODO (Phase 2)
│   └── [tests]
├── utils/
│   ├── errors.ts                               # TODO
│   ├── validators.ts                           # TODO (Phase 2)
│   └── helpers.ts
└── __tests__/
    ├── unit/
    │   ├── config.test.ts                      # ✅ DONE
    │   ├── logger.test.ts                      # ✅ DONE
    │   ├── [remaining unit tests]
    └── integration/
        └── [route integration tests]
```

---

## 📈 Progress

- **Phase 1 Plumbing:** 2/10 (Config, Logger)
- **Phase 2 APIs:** 0/11 (Services, Routes, Events)
- **Tests:** 14/30+ passing
- **Est. Time to Phase 1 Complete:** 7-8 hours
- **Est. Time to Phase 2 Complete:** 8-10 hours
- **Total Sprint:** 16-18 hours

---

## 🎯 Definition of Done

Phase 1 complete:
- ✅ `pnpm install` succeeds
- ✅ `pnpm test` passes all Phase 1 tests
- ✅ `pnpm run build` compiles clean
- ✅ `SERVICE_PORT=7002 pnpm dev` starts without errors
- ✅ `curl http://localhost:7002/healthz` returns 200 OK
- ✅ `curl http://localhost:7002/metrics` returns Prometheus format

Phase 2 complete:
- ✅ All 15 API endpoints work
- ✅ Cross-tenant access blocked (403)
- ✅ Invalid input rejected (400)
- ✅ NATS events published on mutation
- ✅ Test coverage > 80%
- ✅ `pnpm test` passes 30+ tests

---

## Commands

```bash
# Install dependencies
pnpm install

# Run all tests
pnpm test

# Run specific test file
pnpm test src/__tests__/unit/config.test.ts

# Watch tests
pnpm test --watch

# Build
pnpm run build

# Start dev server (after Phase 1)
SERVICE_PORT=7002 pnpm dev

# Run migrations (after Phase 1)
pnpm migrate

# Seed test data (Phase 2)
pnpm seed
```

---

## Success Metrics

- ✅ **Config:** 7/7 tests passing
- ✅ **Logger:** 7/7 tests passing
- ⏳ **Remaining Phase 1:** 16+ tests to implement
- ⏳ **Phase 2:** 30+ tests to implement

**Current Status:** Phase 1 is 20% complete (2/10 files). Ready to continue with database layer.
