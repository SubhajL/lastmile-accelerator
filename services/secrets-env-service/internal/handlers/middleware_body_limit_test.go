package handlers

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
)

func TestBodyLimit_Rejects_WhenContentLengthExceeds(t *testing.T) {
    r := chi.NewRouter()
    r.Use(BodyLimit(4))
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("0123456789"))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusRequestEntityTooLarge {
        t.Fatalf("expected 413, got %d", w.Code)
    }
}

func TestBodyLimit_Allows_WhenBelowLimit(t *testing.T) {
    r := chi.NewRouter()
    r.Use(BodyLimit(1024))
    r.Post("/echo", okHandler)
    req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("{}"))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestBodyLimit_IgnoresGET(t *testing.T) {
    r := chi.NewRouter()
    r.Use(BodyLimit(4))
    r.Get("/ok", okHandler)
    req := httptest.NewRequest(http.MethodGet, "/ok", bytes.NewBufferString("0123456789"))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }
}
