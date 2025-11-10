package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"example.com/lma/db-guardian-service/internal/dto"
)

// Minimal v1 routes using existing services; auth/ratelimit hooks deferred to middlewares chain
func registerAPIv1(mux *http.ServeMux, _ any, deps *Dependencies) {
	// POST /v1/projects/{id}/db/connections
	mux.HandleFunc("/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/v1/projects/") {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.TrimPrefix(path, "/v1/projects/"), "/")
		if len(parts) < 3 { http.NotFound(w, r); return }
		projectID := parts[0]
		resource := strings.Join(parts[1:], "/")

		switch r.Method + " " + resource {
		case http.MethodPost + " db/connections":
			var req dto.RegisterConnectionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
			if req.ProjectID == "" { req.ProjectID = projectID }
			if err := req.Validate(); err != nil || req.ProjectID != projectID { http.Error(w, "invalid payload", http.StatusBadRequest); return }
			m := req.ToModel()
			id, err := deps.ConnSvc.RegisterConnection(r.Context(), &m, req.MakeDefault)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusCreated, dto.ConnectionResponse{ID: id, ProjectID: projectID, Name: m.Name, Driver: m.Driver, DSNRef: m.DSNRef, IsDefault: m.IsDefault})
			return

		case http.MethodGet + " db/connections":
			list, err := deps.ConnSvc.ListConnections(r.Context(), projectID)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, list)
			return

		case http.MethodPost + " db/analyze":
			var req dto.RunAnalysisRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
			if req.ProjectID == "" { req.ProjectID = projectID }
			if err := req.Validate(); err != nil || req.ProjectID != projectID { http.Error(w, "invalid payload", http.StatusBadRequest); return }
			role, val, idx := req.ToOptions(nil)
			rep, err := deps.AnalysisSvc.RunFullAnalysis(r.Context(), projectID, req.MigrationName, req.SQL, role, val, idx)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, dto.NewAnalysisReportResponse(rep.Role, rep.Migration, rep.Index))
			return

		case http.MethodPost + " db/migrations/validate":
			var req dto.ValidateMigrationRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
			if req.ProjectID == "" { req.ProjectID = projectID }
			if err := req.Validate(); err != nil || req.ProjectID != projectID { http.Error(w, "invalid payload", http.StatusBadRequest); return }
			res, err := deps.MigGuard.ValidateMigration(r.Context(), req.SQL, req.ToValidationOptions(nil))
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, dto.NewMigrationValidationResponse(res))
			return

		case http.MethodGet + " db/recommendations":
			onlyUnapplied := r.URL.Query().Get("only_unapplied") == "true"
			idx, role, err := deps.RecsProvider.Get(r.Context(), projectID, onlyUnapplied)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, map[string]any{"index": idx, "role": role})
			return

		case http.MethodGet + " db/policies":
			p, err := deps.PolicyMgr.GetPolicy(r.Context(), projectID)
			if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }
			writeJSON(w, http.StatusOK, dto.PolicyResponse{ProjectID: p.ProjectID, SpecYAML: p.SpecYAML, Version: p.Version})
			return

		case http.MethodPut + " db/policies":
			var body struct{ SpecYAML string `json:"spec_yaml"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
			if strings.TrimSpace(body.SpecYAML) == "" { http.Error(w, "spec_yaml required", http.StatusBadRequest); return }
			p, err := deps.PolicyMgr.UpdatePolicy(r.Context(), projectID, body.SpecYAML)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, dto.PolicyResponse{ProjectID: p.ProjectID, SpecYAML: p.SpecYAML, Version: p.Version})
			return

		case http.MethodGet + " db/drift":
			d, err := deps.DriftCheck.Check(r.Context(), projectID)
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			writeJSON(w, http.StatusOK, d)
			return
		}

		http.NotFound(w, r)
	})
}
