package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/dto"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/service"
)

// Contracts allow injecting fakes in tests without depending on concrete service types.
type ConnectionManager interface {
	RegisterConnection(ctx context.Context, conn *models.DBConnection, makeDefault bool) (string, error)
	DeleteConnection(ctx context.Context, id string) error
	GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error)
	ListConnections(ctx context.Context, projectID string) ([]models.DBConnection, error)
}

type AnalysisRunner interface {
	RunFullAnalysis(ctx context.Context, projectID, migrationName, migrationSQL string, role analyzer.AnalyzeOptions, val analyzer.ValidationOptions, idx analyzer.IndexAnalysisOptions) (*service.AnalysisReport, error)
}

type MigrationValidator interface {
	ValidateMigration(ctx context.Context, sql string, opts analyzer.ValidationOptions) (*analyzer.MigrationValidationResult, error)
}

// New v1 support interfaces
// Policies
type PolicyManager interface {
	GetPolicy(ctx context.Context, projectID string) (*models.RolePolicy, error)
	UpdatePolicy(ctx context.Context, projectID, yaml string) (*models.RolePolicy, error)
}

// Recommendations
type RecommendationsProvider interface {
	Get(ctx context.Context, projectID string, onlyUnapplied bool) ([]analyzer.IndexRecommendation, *analyzer.RoleAnalysisResult, error)
}

// Drift
type DriftChecker interface {
	Check(ctx context.Context, projectID string) (*dto.DriftResponse, error)
}

func registerAPI(mux *http.ServeMux, cfg *config.Config, deps *Dependencies) {
	// Connections
	mux.HandleFunc("/api/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if deps.ConnSvc == nil {
			http.Error(w, "connection service unavailable", http.StatusServiceUnavailable)
			return
		}
		var req dto.RegisterConnectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		m := req.ToModel()
		id, err := deps.ConnSvc.RegisterConnection(r.Context(), &m, req.MakeDefault)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := dto.ConnectionResponse{ID: id, ProjectID: req.ProjectID, Name: req.Name, Driver: req.Driver, DSNRef: req.DSNRef, IsDefault: req.MakeDefault}
		writeJSON(w, http.StatusCreated, resp)
	})

	mux.HandleFunc("/api/connections/", func(w http.ResponseWriter, r *http.Request) {
		if deps.ConnSvc == nil {
			http.Error(w, "connection service unavailable", http.StatusServiceUnavailable)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/connections/")
		if path == "default" && r.Method == http.MethodGet {
			projectID := r.URL.Query().Get("project_id")
			if projectID == "" {
				http.Error(w, "project_id is required", http.StatusBadRequest)
				return
			}
			conn, err := deps.ConnSvc.GetDefaultConnection(r.Context(), projectID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, dto.ConnectionResponse{ID: conn.ID, ProjectID: conn.ProjectID, Name: conn.Name, Driver: conn.Driver, DSNRef: conn.DSNRef, IsDefault: conn.IsDefault})
			return
		}
		if r.Method == http.MethodDelete {
			id := path
			if id == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := deps.ConnSvc.DeleteConnection(r.Context(), id); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	// Migration validation
	mux.HandleFunc("/api/migrations/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if deps.MigGuard == nil {
			http.Error(w, "migration validator unavailable", http.StatusServiceUnavailable)
			return
		}
		var req dto.ValidateMigrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		opts := req.ToValidationOptions(cfg)
		res, err := deps.MigGuard.ValidateMigration(r.Context(), req.SQL, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf("validation failed: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, dto.NewMigrationValidationResponse(res))
	})

	// Full analysis
	mux.HandleFunc("/api/analysis/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if deps.AnalysisSvc == nil {
			http.Error(w, "analysis service unavailable", http.StatusServiceUnavailable)
			return
		}
		var req dto.RunAnalysisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		role, val, idx := req.ToOptions(cfg)
		rep, err := deps.AnalysisSvc.RunFullAnalysis(r.Context(), req.ProjectID, req.MigrationName, req.SQL, role, val, idx)
		if err != nil {
			http.Error(w, fmt.Sprintf("analysis failed: %v", err), http.StatusInternalServerError)
			return
		}
		resp := dto.NewAnalysisReportResponse(rep.Role, rep.Migration, rep.Index)
		writeJSON(w, http.StatusOK, resp)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
