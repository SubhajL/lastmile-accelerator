package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"example.com/lma/observability-service/internal/handlers"
	"example.com/lma/observability-service/internal/logs"
	"example.com/lma/observability-service/internal/middleware"
	"example.com/lma/observability-service/internal/repository"
	"example.com/lma/observability-service/internal/services"
	"example.com/lma/observability-service/internal/traces"
)

type mockTracesClient struct{}

func (m *mockTracesClient) SearchTraceQL(ctx context.Context, query string, start, end time.Time, limit int) ([]traces.TraceSummary, error) {
	return []traces.TraceSummary{{TraceID: "abc123", Service: "test-service"}}, nil
}

func (m *mockTracesClient) GetTrace(ctx context.Context, traceID string) (*traces.TraceDetail, error) {
	return &traces.TraceDetail{TraceID: traceID, Spans: []json.RawMessage{[]byte(`{"span":"mock"}`)}}, nil
}

type mockLogsClient struct{}

func (m *mockLogsClient) QueryRange(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error) {
	return []logs.LogEntry{{Timestamp: time.Now(), Line: "mock log line"}}, nil
}

type mockPromClient struct{}

func (m *mockPromClient) Query(ctx context.Context, promQL string, window time.Duration) (float64, error) {
	return 42.0, nil
}

func TestRouteAuth_MissingToken_401(t *testing.T) {
	// Setup mock services
	tracesClient := &mockTracesClient{}
	logsClient := &mockLogsClient{}
	promClient := &mockPromClient{}
	queriesSvc := services.NewObservabilityQueries(tracesClient, logsClient, promClient)
	handler := handlers.NewQueriesHandler(queriesSvc)

	// Mock JWKS server (unused in this test)
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Test protected endpoint without token
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/demo/traces", nil)
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(handler.SearchTraces)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing bearer token") {
		t.Errorf("expected missing token message, got %s", rec.Body.String())
	}
}

func TestRouteAuth_WrongScope_403(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Create token with wrong scope
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":    "test-issuer",
		"sub":    "test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
		"scopes": []string{"observability:write"}, // has write but need read
	})
	token.Header["kid"] = "test-key"
	tokenString, _ := token.SignedString(privKey)

	tracesClient := &mockTracesClient{}
	logsClient := &mockLogsClient{}
	promClient := &mockPromClient{}
	queriesSvc := services.NewObservabilityQueries(tracesClient, logsClient, promClient)
	handler := handlers.NewQueriesHandler(queriesSvc)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/demo/traces", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(handler.SearchTraces)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "forbidden") {
		t.Errorf("expected forbidden message, got %s", rec.Body.String())
	}
}

func TestRouteAuth_ValidToken_Success(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Create valid token with correct scope
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":    "test-issuer",
		"sub":    "test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
		"scopes": []string{"observability:read"},
	})
	token.Header["kid"] = "test-key"
	tokenString, _ := token.SignedString(privKey)

	tracesClient := &mockTracesClient{}
	logsClient := &mockLogsClient{}
	promClient := &mockPromClient{}
	queriesSvc := services.NewObservabilityQueries(tracesClient, logsClient, promClient)
	handler := handlers.NewQueriesHandler(queriesSvc)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/demo/traces?service=test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(handler.SearchTraces)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result []traces.TraceSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("failed to parse response: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty traces result")
	}
}

func TestRouteAuth_WriteEndpoint_RequiresWriteScope(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Test with read scope on write endpoint - should fail
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":    "test-issuer",
		"sub":    "test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
		"scopes": []string{"observability:read"}, // has read but need write
	})
	token.Header["kid"] = "test-key"
	tokenString, _ := token.SignedString(privKey)

	// Mock SLO handler for write test
	sloRepo := repository.NewSLORepository(nil) // nil db for mock
	promClient := &mockPromClient{}
	sloSvc := services.NewSLOService(sloRepo, nil, promClient) // nil tracer for mock
	sloHandler := handlers.NewSLOHandler(sloSvc)

	body := `{"service_name":"test","type":"availability","target":0.99,"window_seconds":300,"query":"up"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/demo/slos", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:write")(http.HandlerFunc(sloHandler.CreateSLO)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for read scope on write endpoint, got %d", rec.Code)
	}
}

func TestRouteAuth_IngestEndpoint_RequiresIngestScope(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Test with write scope on ingest endpoint - should fail
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":    "test-issuer",
		"sub":    "test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
		"scopes": []string{"observability:write"}, // has write but need ingest
	})
	token.Header["kid"] = "test-key"
	tokenString, _ := token.SignedString(privKey)

	// Mock error handler for ingest test
	errRepo := repository.NewErrorRepository(nil) // nil db for mock
	errSvc := services.NewErrorService(errRepo, nil) // nil tracer for mock
	errHandler := handlers.NewErrorHandler(errSvc)

	body := `{"message":"test error","stack":"line 1\nline 2","title":"Test Error"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/demo/errors/ingest", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:ingest")(http.HandlerFunc(errHandler.Ingest)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for write scope on ingest endpoint, got %d", rec.Code)
	}
}

func TestRouteAuth_MultipleScopes_Success(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nBytes := privKey.PublicKey.N.Bytes()
		eBytes := []byte{1, 0, 1}
		jwks := map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test-key",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		}
		json.NewEncoder(w).Encode(jwks)
	}))
	defer jwksServer.Close()

	verifier := middleware.NewJWKSVerifier(jwksServer.URL, "test-issuer", "", "scopes", nil)

	// Create token with multiple scopes
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":    "test-issuer",
		"sub":    "test-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
		"scopes": "observability:read observability:write observability:ingest", // space-delimited string
	})
	token.Header["kid"] = "test-key"
	tokenString, _ := token.SignedString(privKey)

	tracesClient := &mockTracesClient{}
	logsClient := &mockLogsClient{}
	promClient := &mockPromClient{}
	queriesSvc := services.NewObservabilityQueries(tracesClient, logsClient, promClient)
	handler := handlers.NewQueriesHandler(queriesSvc)

	// Test read endpoint
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/demo/traces", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	protectedHandler := middleware.JWT(verifier)(middleware.RequireScopes("observability:read")(http.HandlerFunc(handler.SearchTraces)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with multi-scope token, got %d", rec.Code)
	}
}