package server

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHTTPScopeResolver_ReadEndpointsRequireRead(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/v1/projects/p1/db/policies", nil)
    sc := HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.read" { t.Fatalf("expected db.read, got %v", sc) }

    req = httptest.NewRequest(http.MethodGet, "/v1/projects/p1/db/recommendations", nil)
    sc = HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.read" { t.Fatalf("expected db.read, got %v", sc) }

    req = httptest.NewRequest(http.MethodGet, "/api/connections/default?project_id=p1", nil)
    sc = HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.read" { t.Fatalf("expected db.read, got %v", sc) }
}

func TestHTTPScopeResolver_WriteEndpointsRequireWrite(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/v1/projects/p1/db/connections", nil)
    sc := HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.write" { t.Fatalf("expected db.write, got %v", sc) }

    req = httptest.NewRequest(http.MethodPut, "/v1/projects/p1/db/policies", nil)
    sc = HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.write" { t.Fatalf("expected db.write, got %v", sc) }

    req = httptest.NewRequest(http.MethodPost, "/api/analysis/run", nil)
    sc = HTTPScopeResolver(req)
    if len(sc) != 1 || sc[0] != "db.write" { t.Fatalf("expected db.write, got %v", sc) }
}

func TestHTTPScopeResolver_PublicPathsBypass(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    if sc := HTTPScopeResolver(req); sc != nil { t.Fatalf("expected bypass (nil), got %v", sc) }
    req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
    if sc := HTTPScopeResolver(req); sc != nil { t.Fatalf("expected bypass (nil), got %v", sc) }
}

type spyAuth struct{ called bool }
func (a *spyAuth) Verify(_ context.Context, _ string, _ []string) (*Claims, error) { a.called = true; return &Claims{Subject:"u"}, nil }

func TestRequireScopesFunc_BypassWhenResolverNil(t *testing.T) {
    spy := &spyAuth{}
    mw := RequireScopesFunc(spy, func(r *http.Request) []string { return nil })
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
    h := mw(next)
    req := httptest.NewRequest(http.MethodGet, "/any", nil)
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)
    if rec.Code != 200 { t.Fatalf("expected 200, got %d", rec.Code) }
    if spy.called { t.Fatalf("expected Verify not called on bypass") }
}

func TestGRPCScopeResolver_MethodsMapCorrectly(t *testing.T) {
    if got := GRPCScopeResolver("/dbguardian.v1.DbGuardianService/GetPolicy"); len(got) != 1 || got[0] != "db.read" { t.Fatalf("GetPolicy -> db.read, got %v", got) }
    if got := GRPCScopeResolver("/dbguardian.v1.DbGuardianService/RunAnalysis"); len(got) != 1 || got[0] != "db.write" { t.Fatalf("RunAnalysis -> db.write, got %v", got) }
    if got := GRPCScopeResolver("/dbguardian.v1.DbGuardianService/RegisterConnection"); got[0] != "db.write" { t.Fatalf("RegisterConnection -> db.write, got %v", got) }
    if got := GRPCScopeResolver("/dbguardian.v1.DbGuardianService/ListConnections"); got[0] != "db.read" { t.Fatalf("ListConnections -> db.read, got %v", got) }
}
