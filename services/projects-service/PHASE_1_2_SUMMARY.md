# Phase 1 & 2: Quick Reference

## Overview
**Build foundation + core APIs** for projects-service using **TDD-first** with Fastify + PostgreSQL + NATS + OTel.

---

## Phase 1: Foundation (Plumbing)

### Files to Create

| File | Purpose | Key Functions |
|------|---------|---------------|
| `src/config.ts` | Env var loading | `loadConfig()`, `validateEnv()` |
| `src/logger.ts` | Pino logger setup | `createLogger()`, `getLogger()` |
| `src/db.ts` | Postgres pool | `initDb()`, `getDb()`, `query()`, `transaction()` |
| `src/db/migrations/001_init.sql` | Schema (6 tables) | — |
| `src/db/migrations/migrate.ts` | Migration runner | `runMigrations()` |
| `src/nats.ts` | NATS publisher | `initNats()`, `getNats()`, `publish()` |
| `src/otel.ts` | OTel SDK setup | `initOtel()`, `getTracer()`, `getMeter()` |
| `src/middleware/auth.ts` | JWT auth | `verifyJwt()`, `authMiddleware()`, `extractTenantId()` |
| `src/middleware/errorHandler.ts` | Global errors | `errorHandler()` |
| `src/middleware/requestId.ts` | Request tracing | `requestIdMiddleware()` |
| `src/utils/errors.ts` | Custom errors | `ValidationError`, `AuthError`, `NotFoundError`, `ConflictError` |
| `src/routes/health.ts` | Health + metrics | `registerHealthRoutes()` |
| `src/index.ts` | Main entry (rewrite) | Bootstrap server, register routes, graceful shutdown |

### Phase 1 Test Cases

| File | Test Cases |
|------|-----------|
| `config.test.ts` | Missing env var throws, valid config passes |
| `logger.test.ts` | Logs are JSON, include service.name |
| `db.test.ts` | Pool connects, query executes, transaction rolls back on error |
| `nats.test.ts` | Connects to NATS, publishes message, adds traceparent |
| `otel.test.ts` | Tracer initialized, spans exported, service.name tagged |
| `auth.test.ts` | Valid JWT accepted, invalid rejected, tenant_id extracted |
| `errorHandler.test.ts` | 400 for validation, 401 for auth, 404 for not found |
| `requestId.test.ts` | Request ID added to header, propagated to logs |
| `health.routes.test.ts` | /healthz responds 200, /metrics has Prometheus headers |

---

## Phase 2: Core APIs (CRUD Endpoints)

### Database Schema (6 Tables)

```
tenants (id, name, slug, created_at, updated_at, deleted_at)
projects (id, tenant_id, name, description, created_at, updated_at, deleted_at)
members (id, tenant_id, user_id, email, role, created_at, updated_at)
environments (id, project_id, name, config, created_at, updated_at)
ingestion_modes (id, project_id, mode, is_default, created_at)
```

### Files to Create

| File | Purpose | Key Functions |
|------|---------|---------------|
| `src/db/schema.ts` | TS types | `Tenant`, `Project`, `Member`, `Environment`, `IngestionMode` |
| `src/utils/validators.ts` | Zod schemas | `CreateProjectInput`, `UpdateProjectInput`, `CreateMemberInput`, etc. |
| `src/services/projectService.ts` | Business logic | `listProjects()`, `getProject()`, `createProject()`, `updateProject()`, `deleteProject()` |
| `src/services/tenantService.ts` | Business logic | `getTenant()`, `listTenantMembers()` |
| `src/services/memberService.ts` | Business logic | `addMember()`, `getMember()`, `updateMemberRole()`, `removeMember()` |
| `src/services/environmentService.ts` | Business logic | `listEnvironments()`, `getEnvironment()`, `createEnvironment()`, `updateEnvironment()`, `deleteEnvironment()`, `setIngestionModes()`, `getIngestionModes()` |
| `src/events/eventPublisher.ts` | Event publishing | `publishProjectEvent()`, `publishMemberEvent()`, `publishEnvironmentEvent()` |
| `src/routes/projects.ts` | API routes | `registerProjectRoutes()` (CRUD endpoints) |
| `src/routes/tenants.ts` | API routes | `registerTenantRoutes()` (GET tenant + list members) |
| `src/routes/members.ts` | API routes | `registerMemberRoutes()` (add, update, delete) |
| `src/routes/environments.ts` | API routes | `registerEnvironmentRoutes()` (CRUD + ingestion modes) |

### Phase 2 Test Cases

| File | Test Cases |
|------|-----------|
| `validators.test.ts` | Valid input passes, invalid input throws ZodError |
| `projectService.test.ts` | createProject creates project with default environments; getProject returns correct tenant; getProject throws 404 for wrong tenant; listProjects returns only non-deleted; updateProject publishes event; deleteProject soft-deletes |
| `tenantService.test.ts` | getTenant returns tenant; getTenant throws 404 for deleted; listTenantMembers returns all |
| `memberService.test.ts` | addMember creates and publishes; addMember throws if not owner; updateMemberRole updates; removeMember prevents last owner; getMember throws 404 for wrong tenant |
| `environmentService.test.ts` | createEnvironment creates and publishes; listEnvironments returns only project envs; setIngestionModes updates; deleteEnvironment prevents last env; updateEnvironment validates tenant+project |
| `eventPublisher.test.ts` | publishProjectEvent publishes to topic; events include traceparent; handles NATS errors |
| `projects.routes.test.ts` | GET /v1/projects returns 200; POST creates 201; GET {id} returns 200; PUT updates 200; DELETE returns 204; GET from other tenant returns 403; invalid input returns 400 |
| `tenants.routes.test.ts` | GET /v1/tenants/{id} returns 200; GET from other tenant returns 403; GET /v1/tenants/{id}/members returns list |
| `members.routes.test.ts` | POST /v1/tenants/{id}/members creates 201; POST by non-owner returns 403; PUT updates 200; DELETE returns 204; DELETE last owner returns 409 |
| `environments.routes.test.ts` | GET /v1/projects/{id}/environments returns 200; POST creates 201; PUT updates 200; DELETE returns 204; DELETE last env returns 409; POST ingestion-modes sets modes |

### Fixtures for Testing

```
fixtures/tenants.fixture.ts     → createTestTenant()
fixtures/projects.fixture.ts    → createTestProject()
fixtures/members.fixture.ts     → createTestMember()
setup.ts                        → withDbTransaction(), globalSetup(), globalTeardown()
```

---

## API Endpoints (Phase 2)

### Projects
- `GET /v1/projects` → List
- `POST /v1/projects` → Create
- `GET /v1/projects/{projectId}` → Get
- `PUT /v1/projects/{projectId}` → Update
- `DELETE /v1/projects/{projectId}` → Delete

### Tenants
- `GET /v1/tenants/{tenantId}` → Get
- `GET /v1/tenants/{tenantId}/members` → List members

### Members
- `POST /v1/tenants/{tenantId}/members` → Add
- `PUT /v1/members/{memberId}` → Update role
- `DELETE /v1/members/{memberId}` → Remove

### Environments
- `GET /v1/projects/{projectId}/environments` → List
- `POST /v1/projects/{projectId}/environments` → Create
- `GET /v1/projects/{projectId}/environments/{envId}` → Get
- `PUT /v1/projects/{projectId}/environments/{envId}` → Update
- `DELETE /v1/projects/{projectId}/environments/{envId}` → Delete
- `GET /v1/projects/{projectId}/ingestion-modes` → List modes
- `POST /v1/projects/{projectId}/ingestion-modes` → Set modes

---

## Key Design Principles

1. **Tenant Isolation**: All queries filter by `tenant_id`; enforced via FK at DB layer
2. **TDD-First**: Write tests before code; green tests gate features
3. **Event-Driven**: Every mutation publishes NATS event for downstream services
4. **Transaction Safety**: Multi-step ops wrapped in DB transactions
5. **Type Safety**: Strict TS + Zod validation for all inputs
6. **Observability**: OTel spans + Pino logs + Prometheus metrics

---

## Implementation Checklist

### Phase 1
- [ ] Update `package.json` + `tsconfig.json`
- [ ] `config.ts` + test
- [ ] `logger.ts` + test
- [ ] `db.ts` + migration + test
- [ ] `nats.ts` + test
- [ ] `otel.ts` + test
- [ ] `middleware/auth.ts` + test
- [ ] `middleware/errorHandler.ts` + test
- [ ] `middleware/requestId.ts` + test
- [ ] `routes/health.ts` + test
- [ ] `utils/errors.ts`
- [ ] `index.ts` (bootstrap)
- [ ] ✅ `pnpm dev` works, `/healthz` responds

### Phase 2
- [ ] `db/schema.ts`
- [ ] `utils/validators.ts` + test
- [ ] Fixtures (tenants, projects, members, setup)
- [ ] `services/projectService.ts` + test
- [ ] `services/tenantService.ts` + test
- [ ] `services/memberService.ts` + test
- [ ] `services/environmentService.ts` + test
- [ ] `events/eventPublisher.ts` + test
- [ ] `routes/projects.ts` + test
- [ ] `routes/tenants.ts` + test
- [ ] `routes/members.ts` + test
- [ ] `routes/environments.ts` + test
- [ ] ✅ All tests pass, coverage > 80%

---

## Success Metrics

- ✅ All tests pass (`pnpm test`)
- ✅ Build succeeds (`npm run build`)
- ✅ Service starts (`SERVICE_PORT=7002 pnpm dev`)
- ✅ Health check responds (`curl http://localhost:7002/healthz`)
- ✅ Metrics exposed (`curl http://localhost:7002/metrics`)
- ✅ Test coverage > 80%
