# Secrets Env Service — MVP Hardening Progress Tracker

Last Updated: 2025-11-12
Specification: ./secrets-env-service-mvp-hardening-spec.md

## Overview
Track completion of MVP productionization work across traces, metrics, validation/security, operability, and docs.

## Phase Completion Summary
| Phase                | Status | Completion | Notes |
|---------------------:|:------:|:----------:|-------|
| Traces (OTLP)        |   ⏳   |   0%       | Add OTLP exporter and shutdown |
| Metrics              |   ⏳   |   0%       | HTTP/gRPC/domain counters & histograms |
| Validation/Security  |   ⏳   |   0%       | Env allowlist, content-type, size cap, RBAC claims |
| Operability          |   ⏳   |   0%       | gRPC reflection/health; startup readiness |
| Config & Docs        |   ⏳   |   0%       | New env vars and README updates |

## Current Tasks (Critical Path)
- [ ] Wire OTLP trace exporter (gRPC) with config toggles; verify traces land in collector.
- [ ] Expose application metrics: http_requests_total, http_request_duration_seconds; grpc_server_handled_total, grpc_server_handling_seconds; domain counters.
- [ ] Enforce env allowlist {dev,staging,prod} (configurable); reject unknown Content-Type; cap request body size.
- [ ] Prefer RBAC roles from JWT claims; keep header override for dev/tests.
- [ ] Enable gRPC health service and reflection under a config toggle.
- [ ] Startup readiness checks: DB fail-fast; Redis/NATS degraded with warnings; Vault health check (non-test).
- [ ] Update service README with new env vars and example curl/grpcurl commands.

## What’s Next
Implement phases in order: Traces → Metrics → Validation/Security → Operability → Docs. Keep all existing tests green.

## Blockers/Issues
None known.
