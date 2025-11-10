package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"example.com/lma/observability-service/internal/httpjson"
	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/services"
)

type SLOHandler struct{ service services.SLOService }

func NewSLOHandler(s services.SLOService) *SLOHandler { return &SLOHandler{service: s} }

// CreateSLO expects body: { service_name, type, target, window_seconds, query }
func (h *SLOHandler) CreateSLO(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
	var body struct {
		ServiceName   string  `json:"service_name"`
		Type          string  `json:"type"`
		Target        float64 `json:"target"`
		WindowSeconds int64   `json:"window_seconds"`
		Query         string  `json:"query"`
	}
	if err := httpjson.StrictDecode(w, r, &body, httpjson.DefaultMaxBody); err != nil {
		if err == httpjson.ErrTooLarge { h.writeError(w, http.StatusRequestEntityTooLarge, "payload too large"); return }
		h.writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	slo := &models.SLO{ProjectID: projectID, ServiceName: body.ServiceName, Type: models.SLOType(body.Type), Target: body.Target, Window: time.Duration(body.WindowSeconds) * time.Second, Query: body.Query}
	if err := h.service.CreateSLO(r.Context(), slo); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, slo)
}

func (h *SLOHandler) GetSLO(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	slo, err := h.service.GetSLO(r.Context(), id)
	if err != nil || slo == nil {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	h.writeJSON(w, http.StatusOK, slo)
}

func (h *SLOHandler) ListProjectSLOs(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
	slos, err := h.service.ListProjectSLOs(r.Context(), projectID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, slos)
}

func (h *SLOHandler) UpdateSLO(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	var body struct {
		ServiceName   string  `json:"service_name"`
		Type          string  `json:"type"`
		Target        float64 `json:"target"`
		WindowSeconds int64   `json:"window_seconds"`
		Query         string  `json:"query"`
	}
	if err := httpjson.StrictDecode(w, r, &body, httpjson.DefaultMaxBody); err != nil {
		if err == httpjson.ErrTooLarge { h.writeError(w, http.StatusRequestEntityTooLarge, "payload too large"); return }
		h.writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	slo := &models.SLO{ID: id, ServiceName: body.ServiceName, Type: models.SLOType(body.Type), Target: body.Target, Window: time.Duration(body.WindowSeconds) * time.Second, Query: body.Query}
	if err := h.service.UpdateSLO(r.Context(), slo); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, slo)
}

func (h *SLOHandler) DeleteSLO(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	if err := h.service.DeleteSLO(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SLOHandler) GetSLOStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-2]
	status, err := h.service.GetSLOStatus(r.Context(), id)
	if err != nil || status == nil {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	h.writeJSON(w, http.StatusOK, status)
}

func (h *SLOHandler) GetSLOHistory(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-2]
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)
	rows, err := h.service.GetSLOHistory(r.Context(), id, from, to)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, rows)
}

func (h *SLOHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *SLOHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
