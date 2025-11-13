# Authentication Enforcement (JWT/JWKS) — Technical Specification

Document Name: Authentication Enforcement Implementation Plan
Date: 2025-11-12
Version: 1.0
Status: Active

## Executive Summary
Add JWT/JWKS middleware to enforce authentication on all `/v1/**` routes. Validate signature via JWKS, verify `exp`, `iss`, `aud`, and extract `tenant_id`, `sub`, and `scopes` into request extensions.

## Architecture Overview
- Location: `services/dep-governance-service/src/middleware/auth.rs`
- Inputs: `Authorization: Bearer <JWT>`; JWKS URL from `JWT_PUBLIC_KEY_URL` env.
- Components: JWKS fetcher with caching and periodic refresh; validator (alg=RS256/ES256 per keys);
  Axum layer to inject `AuthClaims` into extensions.
- Apply layer to `/v1` router only.

## Implementation Phases
1. Types: `AuthClaims { tenant_id: Uuid, sub: String, scopes: Vec<String>, ... }`.
2. JWKS client: fetch, cache, kid-based key lookup, background refresh with jitter.
3. Validator: decode/verify, check `exp`, `iss`, `aud` against config.
4. Axum middleware: reject with 401 on missing/invalid; attach claims on success.
5. Wire into `main.rs` for `/v1` subtree.

## Testing & Verification
- Unit tests for validator (valid, bad sig, expired, wrong iss/aud).
- HTTP-style tests in `tests/middleware_auth_http_test.rs` against a minimal router:
  valid -> 200; bad/expired -> 401; missing -> 401.

## Security Considerations
- Enforce minimal accepted algorithms; ignore `none`.
- Timeout/retry on JWKS fetch; fail-closed if unreachable (configurable grace window).
- Cache TTL and key rotation handled via `kid` and refresh loop.
