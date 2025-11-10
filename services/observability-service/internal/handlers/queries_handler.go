package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/lma/observability-service/internal/services"
)

type QueriesHandler struct{ svc *services.ObservabilityQueries }

func NewQueriesHandler(s *services.ObservabilityQueries) *QueriesHandler { return &QueriesHandler{svc: s} }

// GET /v1/projects/{id}/traces
func (h *QueriesHandler) SearchTraces(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	service := q.Get("service")
	operation := q.Get("operation")
	minDur, _ := time.ParseDuration(emptyToZero(q.Get("minDuration")))
	maxDur, _ := time.ParseDuration(emptyToZero(q.Get("maxDuration")))
	errorOnly := q.Get("errorOnly") == "true"
	userQ := q.Get("query")
	limit, _ := strconv.Atoi(emptyToZero(q.Get("limit")))
	start, _ := time.Parse(time.RFC3339, q.Get("start"))
	end, _ := time.Parse(time.RFC3339, q.Get("end"))
	res, err := h.svc.SearchTraces(r.Context(), service, operation, minDur, maxDur, errorOnly, userQ, start, end, limit)
	if err != nil { writeErr(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, res)
}

// GET /v1/traces/{traceId}
func (h *QueriesHandler) GetTrace(w http.ResponseWriter, r *http.Request) {
	id := lastSeg(r.URL.Path)
	res, err := h.svc.GetTrace(r.Context(), id)
	if err != nil || res == nil { writeErr(w, http.StatusNotFound, "not found"); return }
	writeJSON(w, http.StatusOK, res)
}

// GET /v1/projects/{id}/logs
func (h *QueriesHandler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	logQL := q.Get("q")
	if logQL == "" { writeErr(w, http.StatusBadRequest, "q required"); return }
	limit, _ := strconv.Atoi(emptyToZero(q.Get("limit")))
	direction := q.Get("direction")
	start, _ := time.Parse(time.RFC3339, q.Get("start"))
	end, _ := time.Parse(time.RFC3339, q.Get("end"))
	res, err := h.svc.SearchLogs(r.Context(), logQL, start, end, limit, direction)
	if err != nil { writeErr(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, res)
}

// GET /v1/projects/{id}/dashboards/golden
func (h *QueriesHandler) GoldenDashboard(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	service := q.Get("service")
	windowStr := q.Get("window")
	if windowStr == "" { windowStr = "5m" }
	window, err := time.ParseDuration(windowStr)
	if err != nil { writeErr(w, http.StatusBadRequest, "invalid window"); return }
	res, err := h.svc.Golden(r.Context(), service, window)
	if err != nil { writeErr(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, res)
}

func lastSeg(p string) string { parts := strings.Split(strings.Trim(p, "/"), "/"); return parts[len(parts)-1] }
func emptyToZero(s string) string { if s == "" { return "0" }; return s }
func writeJSON(w http.ResponseWriter, status int, v any){ w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(v)}
func writeErr(w http.ResponseWriter, status int, msg string){ writeJSON(w, status, map[string]string{"error": msg}) }