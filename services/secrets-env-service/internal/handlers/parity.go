package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/go-chi/chi/v5"
)

type parityServicePort interface {
	CheckParity(ctx any, projectID, baseEnv, compareEnv string) (*domain.EnvParityCheck, error)
	GetLatestCheck(ctx any, projectID string) (*domain.EnvParityCheck, error)
	GetCheckHistory(ctx any, projectID string, limit int) ([]*domain.EnvParityCheck, error)
}

type ParityHandler struct { svc parityServicePort }

func NewParityHandler(svc parityServicePort) *ParityHandler { return &ParityHandler{svc: svc} }

type parityReq struct { BaseEnv string `json:"baseEnv"`; CompareEnv string `json:"compareEnv"` }

func (h *ParityHandler) CheckParity(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	var req parityReq
	dec := json.NewDecoder(r.Body); dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil { Error(w, http.StatusBadRequest, "invalid json", err); return }
	if req.BaseEnv == "" || req.CompareEnv == "" { ValidationError(w, map[string]string{"baseEnv":"required","compareEnv":"required"}); return }
	res, err := h.svc.CheckParity(r.Context(), projectID, req.BaseEnv, req.CompareEnv)
	if err != nil { Error(w, http.StatusInternalServerError, "failed parity", err); return }
	Success(w, http.StatusOK, res)
}

func (h *ParityHandler) GetLatestCheck(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	res, err := h.svc.GetLatestCheck(r.Context(), projectID)
	if err != nil { Error(w, http.StatusNotFound, "not found", err); return }
	Success(w, http.StatusOK, res)
}

func (h *ParityHandler) GetCheckHistory(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" { if n, err := strconv.Atoi(s); err == nil { limit = n } }
	res, err := h.svc.GetCheckHistory(r.Context(), projectID, limit)
	if err != nil { Error(w, http.StatusInternalServerError, "failed", err); return }
	Success(w, http.StatusOK, res)
}
