# Error Mapping Polish — Technical Specification

Document Name: Database Error Mapping Plan
Date: 2025-11-12
Version: 1.0
Status: Active

## Executive Summary
Introduce a helper to consistently translate `sqlx::Error` to `AppError`, mapping Postgres codes 23505 (unique violation) and 23503 (foreign key violation) to appropriate 409/400 responses.

## Architecture Overview
- Helper: `src/errors/db.rs` -> `pub fn map_db_error(e: sqlx::Error) -> AppError`
- Integrate across handlers that persist SBOMs, CVEs, vulnerability links, etc.

## Implementation Phases
1. Implement mapper reading `pg_error.code()`.
2. Replace ad-hoc mappings in handlers with `map_db_error`.
3. Add tests for duplicate SBOM `storage_key` -> 409; FK violation -> 400/409 per context.

## Testing & Verification
- Unit tests for mapper; integration tests that hit DB unique/FK constraints.

## Security Considerations
- Do not leak raw DB messages; sanitize hints in response.
