# Projects Service API

This service provides REST endpoints for managing projects, tenants, members, and environments.

Security
- JWT bearer auth (RS256 via JWKS)
- Scopes per endpoint:
  - projects:read — list/get projects
  - projects:write — create/update/delete projects
  - tenants:read — get tenant info and members
  - members:write — add/update/remove members
  - environments:read — list environments
  - environments:write — create/update/delete environments, set ingestion modes

OpenAPI
- See `docs/openapi.yaml` (OpenAPI 3.1)

Errors
- Standard shape:
  - code: string
  - message: string
  - details: object (optional)

Observability
- OpenTelemetry tracing with auto-instrumentations for HTTP and DB; spans include attributes: http.route, http.request_id, enduser.id, tenant.id, http.status_code
- Events published to NATS carry W3C traceparent from the active span
- /metrics exposes Prometheus text with totals and duration summaries; OTel histogram `http.server.duration` records per-request durations
