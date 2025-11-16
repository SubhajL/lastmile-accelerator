package server

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "example.com/lma/secrets-env-service/internal/handlers"
    "example.com/lma/secrets-env-service/internal/startup"
    appLogger "example.com/lma/secrets-env-service/internal/logger"
)

func TestReadyz_Returns503_WhenNotReady(t *testing.T) {
    startup.SetReady(false)
    r := SetupRoutes(nil, nil, nil, handlers.PanicRecovery(appLogger.New("test","info", nil)))
    req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusServiceUnavailable { t.Fatalf("expected 503, got %d", w.Code) }
}

func TestReadyz_Returns200_WhenReady(t *testing.T) {
    startup.SetReady(true)
    r := SetupRoutes(nil, nil, nil, handlers.PanicRecovery(appLogger.New("test","info", nil)))
    req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}
