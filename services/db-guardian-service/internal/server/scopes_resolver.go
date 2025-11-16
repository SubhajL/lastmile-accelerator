package server

import (
    "net/http"
    "strings"
)

// HTTPScopeResolver maps HTTP requests to required scopes.
// Returns nil to explicitly bypass auth for public endpoints.
func HTTPScopeResolver(r *http.Request) []string {
    path := r.URL.Path
    // Public
    if path == "/healthz" || path == "/metrics" { return nil }

    // v1 routes under projects
    if strings.HasPrefix(path, "/v1/projects/") {
        // resource is whatever after /v1/projects/{id}/
        parts := strings.Split(strings.TrimPrefix(path, "/v1/projects/"), "/")
        if len(parts) >= 2 {
            resource := strings.Join(parts[1:], "/")
            switch r.Method + " " + resource {
            case http.MethodGet + " db/connections",
                 http.MethodGet + " db/recommendations",
                 http.MethodGet + " db/policies",
                 http.MethodGet + " db/drift":
                return []string{"db.read"}
            case http.MethodPost + " db/connections",
                 http.MethodPost + " db/analyze",
                 http.MethodPost + " db/migrations/validate",
                 http.MethodPut + " db/policies":
                return []string{"db.write"}
            }
        }
        // Default heuristic for known area
        if r.Method == http.MethodGet { return []string{"db.read"} }
        if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete { return []string{"db.write"} }
    }

    // legacy /api
    if strings.HasPrefix(path, "/api/") {
        if path == "/api/connections" && r.Method == http.MethodPost { return []string{"db.write"} }
        if strings.HasPrefix(path, "/api/connections/") && r.Method == http.MethodGet { return []string{"db.read"} }
        if path == "/api/migrations/validate" && r.Method == http.MethodPost { return []string{"db.write"} }
        if path == "/api/analysis/run" && r.Method == http.MethodPost { return []string{"db.write"} }
        if r.Method == http.MethodGet { return []string{"db.read"} }
        return []string{"db.write"}
    }

    // Unknown paths: require read for GET, write otherwise (conservative)
    if r.Method == http.MethodGet { return []string{"db.read"} }
    return []string{"db.write"}
}

// GRPCScopeResolver maps gRPC full method name to required scopes.
func GRPCScopeResolver(method string) []string {
    switch {
    case strings.Contains(method, "/RegisterConnection"),
         strings.Contains(method, "/UpdatePolicy"),
         strings.Contains(method, "/ValidateMigration"),
         strings.Contains(method, "/RunAnalysis"):
        return []string{"db.write"}
    case strings.Contains(method, "/ListConnections"),
         strings.Contains(method, "/GetRecommendations"),
         strings.Contains(method, "/GetPolicy"),
         strings.Contains(method, "/CheckDrift"):
        return []string{"db.read"}
    }
    return []string{"db.read"}
}
