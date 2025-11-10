package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/lma/observability-service/internal/httpjson"
	"example.com/lma/observability-service/internal/services"
	"example.com/lma/observability-service/internal/models"
)

type ErrorHandler struct{ svc services.ErrorService }

func NewErrorHandler(s services.ErrorService) *ErrorHandler { return &ErrorHandler{svc: s} }

func (h *ErrorHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
	var body struct{
		Message string `json:"message"`
		Stack string `json:"stack"`
		Fingerprint string `json:"fingerprint"`
		Title string `json:"title"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := httpjson.StrictDecode(w, r, &body, httpjson.ErrorIngestMaxBody); err != nil { if err==httpjson.ErrTooLarge { h.err(w, http.StatusRequestEntityTooLarge, "payload too large"); return }; h.err(w, http.StatusBadRequest, "invalid json"); return }
	gid, err := h.svc.Ingest(r.Context(), projectID, body.Message, body.Stack, body.Fingerprint, body.Title, body.Metadata)
	if err != nil { h.err(w, http.StatusInternalServerError, err.Error()); return }
	h.ok(w, http.StatusCreated, map[string]string{"group_id": gid})
}

func (h *ErrorHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
	gs, err := h.svc.ListGroups(r.Context(), projectID, models.ErrorGroupFilter{})
	if err != nil { h.err(w, http.StatusInternalServerError, err.Error()); return }
	h.ok(w, http.StatusOK, gs)
}

func (h *ErrorHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id := last(r.URL.Path)
	g, err := h.svc.GetGroup(r.Context(), id)
	if err != nil || g == nil { h.err(w, http.StatusNotFound, "not found"); return }
	h.ok(w, http.StatusOK, g)
}

func (h *ErrorHandler) ListGroupEvents(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-2]
	limit := 50; offset := 0
	if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err==nil { limit = n } }
	if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err==nil { offset = n } }
	events, err := h.svc.ListGroupEvents(r.Context(), id, limit, offset)
	if err != nil { h.err(w, http.StatusInternalServerError, err.Error()); return }
	h.ok(w, http.StatusOK, events)
}

func (h *ErrorHandler) ResolveGroup(w http.ResponseWriter, r *http.Request) {
	id := last(strings.TrimSuffix(r.URL.Path, "/resolve"))
	if err := h.svc.ResolveGroup(r.Context(), id); err != nil { h.err(w, http.StatusInternalServerError, err.Error()); return }
	h.ok(w, http.StatusOK, map[string]string{"status":"resolved"})
}

func (h *ErrorHandler) ok(w http.ResponseWriter, status int, v interface{}) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(v) }
func (h *ErrorHandler) err(w http.ResponseWriter, status int, msg string) { h.ok(w, status, map[string]string{"error": msg}) }
func last(p string) string { parts := strings.Split(strings.Trim(p, "/"), "/"); return parts[len(parts)-1] }
