package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/go-chi/chi/v5"
)

type leakScanServicePort interface {
	ScanSnapshot(ctx any, projectID, snapshotID string) ([]*domain.ClientLeakScan, error)
	GetScanResults(ctx any, projectID, snapshotID, severity string) ([]*domain.ClientLeakScan, error)
	MarkAsFixed(ctx any, scanID string) error
}

type LeakScanHandler struct { svc leakScanServicePort }

func NewLeakScanHandler(svc leakScanServicePort) *LeakScanHandler { return &LeakScanHandler{svc: svc} }

type scanReq struct { SnapshotID string `json:"snapshotID"` }

func (h *LeakScanHandler) ScanSnapshot(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	var req scanReq
	dec := json.NewDecoder(r.Body); dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || req.SnapshotID == "" { ValidationError(w, map[string]string{"snapshotID":"required"}); return }
	res, err := h.svc.ScanSnapshot(r.Context(), projectID, req.SnapshotID)
	if err != nil { Error(w, http.StatusInternalServerError, "scan failed", err); return }
	Success(w, http.StatusOK, map[string]any{"count": len(res)})
}

func (h *LeakScanHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	snapshotID := chi.URLParam(r, "snapshotID")
	severity := r.URL.Query().Get("severity")
	res, err := h.svc.GetScanResults(r.Context(), projectID, snapshotID, severity)
	if err != nil { Error(w, http.StatusInternalServerError, "failed", err); return }
	Success(w, http.StatusOK, res)
}

func (h *LeakScanHandler) MarkAsFixed(w http.ResponseWriter, r *http.Request) {
	scanID := chi.URLParam(r, "scanID")
	if scanID == "" { ValidationError(w, map[string]string{"scanID":"required"}); return }
	if err := h.svc.MarkAsFixed(r.Context(), scanID); err != nil { Error(w, http.StatusInternalServerError, "failed", err); return }
	Success(w, http.StatusOK, map[string]string{"status":"ok"})
}
