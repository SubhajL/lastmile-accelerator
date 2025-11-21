# Observability Service — Progress

Last Updated: 2025-11-21  
Specification: ./observability-service-spec.md  
Status: 🚧 Not yet MVP (blockers below)

## Current State
- Foundation in place: config validation, Postgres/Redis clients, health endpoints, OpenTelemetry wiring, JWKS + scope middleware/interceptor, NATS publisher stub, migrations for OTel presets/SLOs/alert rules.
- Domain surfaces implemented: HTTP and gRPC handlers for SLOs, alerts, error inbox, Tempo/Loki/Prom queries, scheduler skeleton with metrics recorder.
- Quality signals: `go test ./...` passing (2025-11-21); proto/doc sync tests enforce env + scope matrices.

## MVP Blockers / Drift vs Spec
- HTTP routing conflict: a second `mux.HandleFunc("/v1/projects/", ...)` overrides the first, so OTel/SLO/traces/logs/golden HTTP routes are unreachable.
- Create flows do not generate IDs (SLOs, alerts, error groups/events) and the error UUID helper returns constant `"0"`, so inserts will fail against UUID PK constraints.
- Error inbox migrations are missing for `error_groups` and `error_events`; ingest/lookup paths will fail at runtime.
- SLO history is never persisted and alert evaluation is not invoked from the scheduler; alert notifications/history remain empty and status data is shallow.
- gRPC create endpoints expect caller-supplied IDs instead of server-generated UUIDs, diverging from contract expectations.

## Phase Completion Snapshot
| Phase | Status | Notes |
| ----- | ------ | ----- |
| Foundation | ✅ | Config, storage, telemetry, health, doc sync tests |
| Domain APIs | ⚠️ | Handlers/services/repos exist but routing + ID/migration gaps block use |
| Operations | ⚠️ | Scheduler runs but no alert wiring/history persistence; CI gates lenient |

## Next Actions
1) Fix the HTTP mux to keep a single `/v1/projects/` handler so OTel/SLO/query routes are reachable.  
2) Generate UUIDs for create paths (SLO/alert/error) and replace the `timeNowUnixNano` stub; update repos to use them.  
3) Add migrations for `error_groups` and `error_events` and extend repository tests to assert statement shapes.  
4) Persist SLO history during evaluations and invoke `AlertService.EvaluateAlerts` from the scheduler; assert alert history writes.  
5) Add end-to-end HTTP+gRPC tests for create/read/status flows to prevent regressions and cover ID/routing issues.

## Documentation & Quality Verification
- Comprehensive CLAUDE.md system integrated with Sabrina Ramonov best practices (BP, C, T, D, O, G rules)
- Custom hooks configured for UserPromptSubmit workflow automation
- Dockerfile Go version aligned with go.mod (golang:1.24)
- Service-specific CLAUDE.md and CONTEXT.md documentation complete

## Test Verification
- `go test ./...` — PASS (2025-11-21)
- All quality gates configured: `make quality` runs vet, lint, test, build
