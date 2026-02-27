package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"example.com/lma/snapshot-orchestrator-service/internal/snapshots"
)

type CreateSnapshotReq struct {
	Mode      snapshots.Mode `json:"mode"`
	SourceRef map[string]any `json:"sourceRef"`
}

func NewRouter(store *snapshots.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Post("/v1/projects/{projectId}/snapshots", func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if projectID == "" {
			http.Error(w, "missing projectId", http.StatusBadRequest)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req CreateSnapshotReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if err := validateMode(req.Mode); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.SourceRef == nil {
			http.Error(w, "missing sourceRef", http.StatusBadRequest)
			return
		}
		if err := validateSourceRef(req.Mode, req.SourceRef); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		snapshot, err := store.Create(r.Context(), projectID, req.Mode, req.SourceRef)
		if err != nil {
			http.Error(w, "failed to create snapshot", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(snapshot)
	})

	r.Get("/v1/snapshots/{snapshotId}", func(w http.ResponseWriter, r *http.Request) {
		snapshotID := chi.URLParam(r, "snapshotId")
		if snapshotID == "" {
			http.Error(w, "missing snapshotId", http.StatusBadRequest)
			return
		}

		snapshot, err := store.Get(r.Context(), snapshotID)
		if err != nil {
			if errors.Is(err, snapshots.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to get snapshot", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snapshot)
	})

	return r
}

func validateMode(mode snapshots.Mode) error {
	switch mode {
	case snapshots.ModeA, snapshots.ModeB, snapshots.ModeC:
		return nil
	default:
		return errors.New("invalid mode")
	}
}

func validateSourceRef(mode snapshots.Mode, sourceRef map[string]any) error {
	if mode != snapshots.ModeC {
		return nil
	}

	zip, ok := sourceRef["zip"]
	if !ok || zip == nil {
		return errors.New("missing sourceRef.zip")
	}
	if _, ok := zip.(map[string]any); !ok {
		return errors.New("invalid sourceRef.zip")
	}
	return nil
}
