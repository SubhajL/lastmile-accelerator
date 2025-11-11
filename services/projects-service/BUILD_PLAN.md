# projects-service Build Plan: Phase 1 & 2

## What We're Building

**projects-service** manages tenants, projects, environments, and team members for the LMA platform. It runs on port `7002` with full multi-tenant isolation, event-driven architecture, and observability.

## Architecture Decision

```
Fastify (HTTP) 
  ↓
PostgreSQL (Data)
  ↓
NATS (Events)
  ↓
OpenTelemetry (Observability)
```

**Testing:** Vitest (fast, ESM-native, TypeScript support)

---

## Phase 1: Foundation (Build the Plumbing)

**Objective:** Get a Fastify server running with all middleware, database, and observability scaffolding so we can start building APIs in Phase 2.

### What's Happening

1. **Setup**: Update `package.json` with 11 prod dependencies (Fastify, pg, nats, OTel) + 8 dev dependencies (vitest, supertest, etc.)
2. **Config**: Read env vars (SERVICE_PORT, DATABASE_URL, NATS_URL, etc.) at startup; validate early
3. **Logging**: Pino logger with JSON output, request IDs, structured logging
4. **Database**: PostgreSQL pool with connection string, retry logic, transaction support
5. **Messaging**: NATS connection with event publishing (includes traceparent for tracing)
6. **Observability**: OpenTelemetry SDK + OTLP exporter, tagged with service.name=projects-service
7. **Middleware**: JWT auth (via Keycloak), global error handler, request ID propagation
8. **Health Check**: `/healthz` endpoint (already exists; verify it works)
9. **Metrics**: `/metrics` endpoint with Prometheus format

### Key Test Coverage

- Config validation (missing vars throw, valid config passes)
- Logger outputs JSON, includes service.name
- DB pool connects, queries execute, transactions rollback on error
- NATS connects and publishes with traceparent
- OTel tracer initializes, spans have service.name tag
- JWT middleware validates tokens and extracts tenant_id
- Error handler returns correct HTTP status codes
- Request ID middleware propagates tracing headers

### Success: Service starts and responds to health check
```bash
SERVICE_PORT=7002 pnpm dev
curl hxxp://localhost:7002/healthz  # {"status": "ok", ...}
curl hxxp://localhost:7002/metrics  # Prometheus format
```

---

## Phase 2: Core APIs (Build the CRUD)

**Objective:** Implement all tenant/project/member/environment CRUD endpoints with full tenant isolation and event publishing.

### Database Schema (6 Tables)

```
tenants         → name, slug, soft-delete
projects        → tenant_id, name, description, soft-delete
members         → tenant_id, user_id, email, role (owner|admin|developer|viewer)
environments    → project_id, name, config (JSONB), 3 per project (dev/staging/prod)
ingestion_modes → project_id, mode (A|B|C), is_default
```

### API Endpoints (15 total)

**Projects (5)**
- `GET /v1/projects` → List user's projects (filtered by tenant_id)
- `POST /v1/projects` → Create project + default environments + add creator as member
- `GET /v1/projects/{projectId}` → Get single project (tenant isolation enforced)
- `PUT /v1/projects/{projectId}` → Update project
- `DELETE /v1/projects/{projectId}` → Soft-delete project

**Tenants (2)**
- `GET /v1/tenants/{tenantId}` → Get tenant
- `GET /v1/tenants/{tenantId}/members` → List all members

**Members (3)**
- `POST /v1/tenants/{tenantId}/members` → Add member (owner-only)
- `PUT /v1/members/{memberId}` → Update role
- `DELETE /v1/members/{memberId}` → Remove member (prevents removing last owner)

**Environments (5)**
- `GET /v1/projects/{projectId}/environments` → List envs
- `POST /v1/projects/{projectId}/environments` → Create env
- `GET /v1/projects/{projectId}/environments/{envId}` → Get env
- `PUT /v1/projects/{projectId}/environments/{envId}` → Update env config
- `DELETE /v1/projects/{projectId}/environments/{envId}` → Delete env (prevents removing last 2)

**Ingestion Modes (2)**
- `GET /v1/projects/{projectId}/ingestion-modes` → List modes for project
- `POST /v1/projects/{projectId}/ingestion-modes` → Set modes + default (for snapshot orchestrator)

### Key Design Decisions

1. **Tenant Isolation**: Every query filters by tenant_id; enforced at both service + DB layer
2. **TDD-First**: Write tests before code; green tests gate each feature
3. **Event-Driven**: Every mutation (create/update/delete) publishes NATS event for downstream listeners
4. **Transaction Safety**: Multi-step operations (e.g., create project + create default envs + add member) wrapped in transactions
5. **Type Safety**: Strict TypeScript + Zod validation for all inputs
6. **Soft Deletes**: Projects/members use deleted_at column for audit trail

### Test Coverage

**30+ test cases across:**
- Unit tests for all services (projectService, tenantService, memberService, environmentService)
- Route tests verifying auth, 403 for cross-tenant access, 400 for invalid input
- Event publisher tests confirming NATS messages include traceparent
- Validator tests confirming Zod schemas work correctly

### Success: All endpoints work with full isolation + observability
```bash
pnpm test                          # All 30+ tests pass
npm run build                      # TypeScript compiles
curl -H "Authorization: Bearer $JWT" hxxp://localhost:7002/v1/projects
# Returns only projects for authenticated user's tenant
```

---

## File Structure

```
src/
├── index.ts                       # Bootstrap server (rewrite)
├── config.ts                      # Env loading + validation
├── logger.ts                      # Pino logger
├── db.ts                          # Postgres pool
├── nats.ts                        # NATS publisher
├── otel.ts                        # OpenTelemetry SDK
├── middleware/
│   ├── auth.ts                    # JWT verification
│   ├── errorHandler.ts            # Global error handler
│   └── requestId.ts               # Request tracing
├── routes/
│   ├── health.ts                  # /healthz + /metrics
│   ├── projects.ts                # Project CRUD
│   ├── tenants.ts                 # Tenant routes
│   ├── members.ts                 # Member routes
│   └── environments.ts            # Environment + ingestion mode routes
├── db/
│   ├── schema.ts                  # TypeScript types
│   └── migrations/
│       ├── 001_init.sql           # Create 6 tables + indexes
│       └── migrate.ts             # Migration runner
├── services/
│   ├── projectService.ts          # Project business logic
│   ├── tenantService.ts           # Tenant business logic
│   ├── memberService.ts           # Member business logic
│   └── environmentService.ts      # Environment business logic
├── events/
│   └── eventPublisher.ts          # Publish NATS events
├── utils/
│   ├── errors.ts                  # Custom error classes
│   ├── validators.ts              # Zod schemas
│   └── helpers.ts                 # Utility functions
└── __tests__/
    ├── setup.ts                   # Test fixtures + DB setup
    ├── fixtures/
    │   ├── tenants.fixture.ts
    │   ├── projects.fixture.ts
    │   └── members.fixture.ts
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
```

---

## Implementation Order (Recommended)

Follow this order to have working code at each step:

1. **package.json** (add dependencies)
2. **config.ts** → test → verify env loading works
3. **logger.ts** → test → verify JSON logging works
4. **db.ts** + migration + test → verify DB pool connects + migrations run
5. **nats.ts** → test → verify messaging works
6. **otel.ts** → test → verify tracing initialized
7. **middleware/auth.ts** → test → verify JWT validation
8. **middleware/errorHandler.ts** → test → verify error handling
9. **middleware/requestId.ts** → test → verify request tracing
10. **routes/health.ts** + test → **CHECKPOINT**: `pnpm dev` works, `/healthz` responds
11. **db/schema.ts** (TypeScript types only, no tests)
12. **utils/validators.ts** → test → verify Zod validation
13. **fixtures/setup.ts** → test DB transaction isolation
14. **services/projectService.ts** → test → all methods work with transactions
15. **services/tenantService.ts** → test
16. **services/memberService.ts** → test
17. **services/environmentService.ts** → test
18. **events/eventPublisher.ts** → test → verify NATS events
19. **routes/projects.ts** + test → route integration tests
20. **routes/tenants.ts** + test
21. **routes/members.ts** + test
22. **routes/environments.ts** + test
23. **index.ts** (register all routes, graceful shutdown)
24. **Polish** → fix coverage, docs, linting

---

## Key Dependencies

### Production
- **fastify** `4.27.0` — HTTP framework
- **pg** `^8.11.3` — PostgreSQL driver
- **nats** `^2.25.0` — NATS messaging
- **@opentelemetry/** — Tracing + metrics
- **pino** `^8.17.2` — Structured logging
- **zod** `^3.22.4` — Input validation
- **jsonwebtoken** — JWT verification
- **uuid** — ID generation

### Development
- **vitest** `^1.1.0` — Fast test runner
- **supertest** `^6.3.3` — HTTP testing
- **typescript** `5.6.3` — Type checking

---

## Success Criteria (GoNoGo)

### Phase 1: ✅ Foundation Complete
- [ ] `pnpm install` succeeds
- [ ] `pnpm dev` starts service on port 7002
- [ ] `curl hxxp://localhost:7002/healthz` returns 200 with `{"status": "ok"}`
- [ ] `curl hxxp://localhost:7002/metrics` returns Prometheus format
- [ ] `pnpm test` passes all 9 Phase 1 tests
- [ ] `npm run build` compiles without errors

### Phase 2: ✅ APIs Complete
- [ ] `pnpm test` passes all 30+ tests (Phase 1 + Phase 2)
- [ ] All 15 endpoints return correct status codes
- [ ] Cross-tenant access returns 403
- [ ] Invalid input returns 400
- [ ] NATS events published on every mutation
- [ ] Test coverage > 80%

---

## Next Steps

1. **Review** this plan with your team
2. **Start Phase 1**: Update package.json and begin with config.ts
3. **Checkpoint**: After health.ts, verify `pnpm dev` works before moving to Phase 2
4. **Phase 2**: Implement services + routes in order
5. **Integration**: Wire into the main LMA platform (snapshot-orchestrator listens to NATS events)

---

## Questions to Ask Yourself

- **Do we have access to test PostgreSQL?** (dev/.env.local has DATABASE_URL=postgres://localhost:55432/lma)
- **Is NATS running?** (dev/.env.local has NATS_URL=nats://localhost:4222)
- **Do we have Keycloak for OIDC?** (dev/.env.local has JWT_JWKS_URL=hxxp://localhost:8080/...)
- **How do we seed test data?** (fixtures + withDbTransaction wrapper in setup.ts)
- **Where do we run migrations?** (migrate.ts runs at service startup, or manually via pnpm migrate)

**All answers in PHASE_1_2_PLAN.md and PHASE_1_2_SUMMARY.md**

---

## Documentation

- **PHASE_1_2_PLAN.md** — Detailed function signatures, test case descriptions, database schema SQL
- **PHASE_1_2_SUMMARY.md** — Quick reference (file table, test table, endpoint list, checklist)
- **This file (BUILD_PLAN.md)** — Executive overview

**Ready to start? Pick a task from Phase 1 and begin with TDD!**
