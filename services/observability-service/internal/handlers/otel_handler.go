package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"example.com/lma/observability-service/internal/httpjson"
	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/services"
)

type OTelHandler struct {
	service services.OTelService
}

func NewOTelHandler(s services.OTelService) *OTelHandler { return &OTelHandler{service: s} }

func (h *OTelHandler) GetPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := h.service.GetAvailablePresets(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, presets)
}

func (h *OTelHandler) GetPresetByFramework(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	fw := models.Framework(parts[len(parts)-1])
	preset, err := h.service.GetPresetForFramework(r.Context(), fw)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "preset not found")
		return
	}
	h.writeJSON(w, http.StatusOK, preset)
}

type applyReq struct {
	Framework string `json:"framework"`
}

func (h *OTelHandler) ApplyConfigToProject(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
var req applyReq
if err := httpjson.StrictDecode(w, r, &req, httpjson.DefaultMaxBody); err != nil {
		if err == httpjson.ErrTooLarge { h.writeError(w, http.StatusRequestEntityTooLarge, "payload too large"); return }
		h.writeError(w, http.StatusBadRequest, "invalid json"); return
}
	cfg, err := h.service.ApplyPresetToProject(r.Context(), projectID, models.Framework(req.Framework))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, cfg)
}

func (h *OTelHandler) GetProjectConfig(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	projectID := parts[2]
	cfg, err := h.service.GetProjectConfiguration(r.Context(), projectID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	h.writeJSON(w, http.StatusOK, cfg)
}

func (h *OTelHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *OTelHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
