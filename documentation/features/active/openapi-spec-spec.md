# Minimal OpenAPI for MVP — Technical Specification

Document Name: OpenAPI (MVP) Specification Plan
Date: 2025-11-12
Version: 1.0
Status: Active

## Executive Summary
Publish an OpenAPI spec documenting current MVP endpoints to enable SDK generation (TS/Go) and consumer integration.

## Architecture Overview
- File: `services/dep-governance-service/openapi.yaml`
- Endpoints included:
  - POST /v1/snapshots/{snapshotId}/sbom
  - GET /v1/snapshots/{snapshotId}/sbom
  - GET /v1/snapshots/{snapshotId}/dependencies
  - GET /v1/dependencies/{dependencyId}/vulns
  - POST /v1/cves
  - POST /v1/dependencies/{dependencyId}/vulns/link
- Schemas mirror existing DTOs used by handlers.

## Implementation Phases
1. Draft OpenAPI with components/schemas and security scheme (Bearer JWT).
2. Validate with `openapi-cli` (or spectral) locally.
3. Add README snippet with curl examples and link spec.

## Testing & Verification
- Lint/validate spec; ensure paths/types match handler I/O.

## Security Considerations
- Define `bearerAuth` and mark secured routes.
