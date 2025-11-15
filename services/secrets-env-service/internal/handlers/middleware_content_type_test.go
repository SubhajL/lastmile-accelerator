package handlers

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
)

func okHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

func TestRequireJSON_AllowsApplicationJSON(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("{}"))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestRequireJSON_AllowsJSONWithCharset(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("{}"))
    req.Header.Set("Content-Type", "application/json; charset=utf-8")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestRequireJSON_BlocksTextPlainWith415(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("x"))
    req.Header.Set("Content-Type", "text/plain")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusUnsupportedMediaType { t.Fatalf("expected 415, got %d", w.Code) }
}

func TestRequireJSON_MissingContentType_WithBody_415(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("{"))
    // no content-type
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusUnsupportedMediaType { t.Fatalf("expected 415, got %d", w.Code) }
}

func TestRequireJSON_IgnoresGETAndDELETE(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Get("/ok", okHandler)
    req := httptest.NewRequest(http.MethodGet, "/ok", nil)
    req.Header.Set("Content-Type", "text/plain")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestRequireJSON_POSTNoBody_Passes(t *testing.T) {
    r := chi.NewRouter()
    r.Use(RequireJSON())
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}
