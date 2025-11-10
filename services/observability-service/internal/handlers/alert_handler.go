package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/lma/observability-service/internal/httpjson"
	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/services"
)

type AlertHandler struct{ service services.AlertService }

func NewAlertHandler(s services.AlertService) *AlertHandler { return &AlertHandler{service: s} }

func (h *AlertHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	sloID := parts[2]
	var body struct {
		Threshold float64  `json:"threshold"`
		Channels  []string `json:"channels"`
		Enabled   bool     `json:"enabled"`
	}
	if err := httpjson.StrictDecode(w, r, &body, httpjson.DefaultMaxBody); err != nil {
		if err == httpjson.ErrTooLarge { h.writeError(w, http.StatusRequestEntityTooLarge, "payload too large"); return }
		h.writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	rule := &models.AlertRule{SLOID: sloID, Threshold: body.Threshold, Enabled: body.Enabled}
	for _, c := range body.Channels {
		rule.Channels = append(rule.Channels, models.AlertChannel(c))
	}
	if err := h.service.CreateAlert(r.Context(), rule); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, rule)
}

func (h *AlertHandler) GetSLOAlerts(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	sloID := parts[2]
	rules, err := h.service.GetSLOAlerts(r.Context(), sloID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, rules)
}

func (h *AlertHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	rule, err := h.service.GetAlert(r.Context(), id)
	if err != nil || rule == nil {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	h.writeJSON(w, http.StatusOK, rule)
}

func (h *AlertHandler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	var body struct {
		Threshold float64  `json:"threshold"`
		Channels  []string `json:"channels"`
		Enabled   bool     `json:"enabled"`
	}
	if err := httpjson.StrictDecode(w, r, &body, httpjson.DefaultMaxBody); err != nil {
		if err == httpjson.ErrTooLarge { h.writeError(w, http.StatusRequestEntityTooLarge, "payload too large"); return }
		h.writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	rule := &models.AlertRule{ID: id, Threshold: body.Threshold, Enabled: body.Enabled}
	for _, c := range body.Channels {
		rule.Channels = append(rule.Channels, models.AlertChannel(c))
	}
	if err := h.service.UpdateAlert(r.Context(), rule); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, rule)
}

func (h *AlertHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
	if err := h.service.DeleteAlert(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-2]
	limit := 100
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			limit = v
		}
	}
	hist, err := h.service.GetAlertHistory(r.Context(), id, limit)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, hist)
}

func (h *AlertHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func (h *AlertHandler) writeError(w http.ResponseWriter, status int, m string) {
	h.writeJSON(w, status, map[string]string{"error": m})
}
