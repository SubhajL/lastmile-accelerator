# Observability Service Configuration Answers

Based on analysis of the LMA codebase and devstack configuration.

---

## 1. Traces/Logs Backends

### ✅ Traces: Tempo
**Base URL:** `hxxp://localhost:3200`

From `dev/.env.local`:
```bash
TEMPO_URL=hxxp://localhost:3200
```

Tempo is running in devstack (`lma-tempo` container) on port 3200. No Jaeger instance found.

### ✅ Logs: Loki
**Base URL:** `hxxp://localhost:3100`

From `dev/.env.local`:
```bash
LOKI_URL=hxxp://localhost:3100
```

Loki is running in devstack (`lma-loki` container) on port 3100.

---

## 2. JWT/JWKS Configuration

### ✅ JWKS URL
```bash
JWKS_URL=hxxp://localhost:8050/realms/lma/protocol/openid-connect/certs
```

**Note:** Keycloak is exposed on port **8050** (not 8080) from `docker-compose.yml`:
```yaml
keycloak:
  ports:
    - "8050:8080"
```

### ✅ JWT Issuer
```bash
JWT_ISSUER=hxxp://localhost:8050/realms/lma
```

### ✅ JWT Audience
Based on existing services, there's **NO specific audience validation** in the current codebase. Services only validate:
- Issuer
- Signature via JWKS
- Expiry/NBF

**Recommendation:** Use empty string or omit `JWT_AUDIENCE` config for now. If needed later, use the service name or a generic value like `lma-platform`.

### ✅ Scope Claim Format
**Claim name:** `scopes` (array or space/comma-delimited string)

Evidence from `secrets-env-service/internal/security/jwtverifier.go`:
```go
func scopesFromClaims(m jwt.MapClaims) []string {
    v, ok := m["scopes"]
    // Handles: []string, []any, or space/comma-delimited string
}
```

And from `test-lab-service/src/middleware/auth.ts`:
```typescript
scopes: payload.scopes || [],
```

**Format:** The codebase accepts:
1. `scopes` as an array: `["scope1", "scope2"]`
2. `scopes` as a string: `"scope1 scope2"` or `"scope1,scope2"`

**Not used:** `scp` or `permissions` claims

---

## 3. Desired Scopes

### ✅ Confirmed Scopes
Based on existing patterns in `observability-service/internal/middleware/scopes_test.go`:

```
observability:read      # GET endpoints (dashboards, queries, trace retrieval)
observability:write     # POST/PUT/DELETE endpoints (SLO config, alert rules)
observability:ingest    # Error/event ingest endpoints
```

These align with patterns seen in other services (e.g., `test-lab:read`, `secrets:write`).

---

## 4. API Shapes and Params

### ✅ Time Parameters
**Format:** RFC3339 timestamps + duration windows

Based on existing service patterns:

```typescript
// Query params for time ranges
start: string       // RFC3339: "2024-01-15T10:00:00Z"
end: string         // RFC3339: "2024-01-15T11:00:00Z"
window: string      // Duration: "5m", "1h", "24h"
```

**Example:**
```
GET /api/v1/metrics/golden-signals?start=2024-01-15T10:00:00Z&end=2024-01-15T11:00:00Z&window=5m
```

### ✅ Trace Search Filters
**Query params:**

```typescript
service: string           // Service name filter
operation: string         // Operation/span name filter
minDuration: string       // e.g., "100ms", "1s"
maxDuration: string       // e.g., "5s"
errorOnly: boolean        // true/false
query: string            // Free-form query (Tempo TraceQL)
limit: number            // Max results (default 20)
```

**Example:**
```
GET /api/v1/traces/search?service=projects-service&errorOnly=true&minDuration=1s&limit=50
```

### ✅ Logs Query
**Query params:**

```typescript
q: string              // Raw LogQL query
limit: number          // Max log lines (default 100)
direction: string      // "forward" or "backward"
start: string          // RFC3339 timestamp
end: string            // RFC3339 timestamp
```

**Example:**
```
GET /api/v1/logs/query?q={service="projects-service"} |= "error"&limit=100&direction=backward
```

---

## 5. Configuration Summary

### Environment Variables for `observability-service`

```bash
# Already in config.go
SERVICE_PORT=7301
GRPC_PORT=50081
PROMETHEUS_URL=hxxp://localhost:9090
OTEL_EXPORTER_OTLP_ENDPOINT=hxxp://localhost:4318

# NEW - Add to config.go
TEMPO_URL=http://localhost:3200
LOKI_URL=http://localhost:3100
JWT_ISSUER=http://localhost:8050/realms/lma
JWT_AUDIENCE=                              # Empty or omit for now
JWT_SCOPE_CLAIM=scopes                     # Claim name for scopes array

# Already configured (correct in .env.local but not in config.go)
JWT_JWKS_URL=http://localhost:8050/realms/lma/protocol/openid-connect/certs
```

**NOTE:** Current `config.go` uses `JWT_PUBLIC_KEY_URL` but should be `JWT_JWKS_URL` for consistency.

---

## 6. Implementation Checklist

### Config Changes
- [ ] Add `TempoURL` to `config.Config` struct
- [ ] Add `LokiURL` to `config.Config` struct
- [ ] Add `JWTIssuer` to `config.Config` struct
- [ ] Add `JWTAudience` to `config.Config` struct (optional)
- [ ] Add `JWTScopeClaim` to `config.Config` struct (default: "scopes")
- [ ] Rename `JWTPublicKeyURL` to `JWTJwksURL` for consistency
- [ ] Update validation logic in `config.Validate()`

### Client Implementations
- [ ] Create `internal/clients/tempo.go` with Tempo HTTP client
  - `SearchTraces(ctx, filters) ([]Trace, error)`
  - `GetTrace(ctx, traceID) (*Trace, error)`
- [ ] Create `internal/clients/loki.go` with Loki HTTP client
  - `QueryLogs(ctx, logQL, limit, direction) ([]LogLine, error)`
- [ ] Add httptest-backed unit tests for both clients

### Service Layer
- [ ] Create `internal/services/observability_queries.go`
  - `GetTraces(ctx, filters) ([]Trace, error)`
  - `GetTrace(ctx, traceID) (*Trace, error)`
  - `QueryLogs(ctx, logQL, params) ([]LogLine, error)`
  - `GetGoldenSignals(ctx, service, window) (*GoldenSignals, error)` (using existing Prom client)

### HTTP Handlers
- [ ] `GET /api/v1/traces/search` - Search traces (scope: `observability:read`)
- [ ] `GET /api/v1/traces/:traceId` - Get single trace (scope: `observability:read`)
- [ ] `GET /api/v1/logs/query` - Query logs (scope: `observability:read`)
- [ ] `GET /api/v1/metrics/golden-signals` - Get RED metrics (scope: `observability:read`)

### Middleware Updates
- [ ] Update JWT middleware to use `JWT_ISSUER` and `JWT_AUDIENCE` for validation
- [ ] Ensure `RequireScopes()` middleware correctly reads from configured scope claim

---

## 7. Devstack Verification

Run these commands to verify devstack services:

```bash
# Check Tempo
curl hxxp://localhost:3200/ready

# Check Loki
curl hxxp://localhost:3100/ready

# Check JWKS endpoint
curl hxxp://localhost:8050/realms/lma/protocol/openid-connect/certs

# Check Prometheus
curl hxxp://localhost:9090/-/healthy
```

**Current Status (from your docker ps):**
- ✅ Tempo: Running
- ✅ Loki: Running
- ✅ Prometheus: Running
- ✅ Keycloak: Running
- ⚠️ OTel Collector: Restarting (check logs if needed)
- ⚠️ Vault: Unhealthy (check logs if needed)

---

## 8. Testing Strategy

### Unit Tests
```go
// internal/clients/tempo_test.go
func TestTempoClient_SearchTraces(t *testing.T) {
    server := httptest.NewServer(...)
    client := NewTempoClient(server.URL)
    // Test search functionality
}
```

### Integration Tests
```bash
# Requires devstack running
cd services/observability-service
export TEMPO_URL=hxxp://localhost:3200
export LOKI_URL=hxxp://localhost:3100
go test -v ./internal/handlers/...
```

---

## References

- **Devstack Config:** `lma-devstack-compose-gitea4001/docker-compose.yml`
- **Dev Environment:** `dev/.env.local`
- **JWT Implementation:** `services/secrets-env-service/internal/security/jwtverifier.go`
- **Scope Middleware:** `services/observability-service/internal/middleware/scopes.go`
- **Existing Config:** `services/observability-service/internal/config/config.go`
