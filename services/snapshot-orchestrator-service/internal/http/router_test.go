package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/snapshot-orchestrator-service/internal/snapshots"
)

func TestCreateSnapshotAndGetSnapshotRoundTrip(t *testing.T) {
	t.Parallel()

	store := snapshots.NewStore()
	router := NewRouter(store)

	createReq := map[string]any{
		"mode": "C",
		"sourceRef": map[string]any{
			"zip": map[string]any{
				"filename":  "repo.zip",
				"sizeBytes": 1,
				"sha256":    "abc",
			},
		},
	}
	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p123/snapshots", bytes.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}

	var created snapshots.Snapshot
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, res.Body.String())
	}
	if created.ProjectID != "p123" {
		t.Fatalf("unexpected projectId: %q", created.ProjectID)
	}
	if created.Mode != snapshots.ModeC {
		t.Fatalf("unexpected mode: %q", created.Mode)
	}
	if created.Status != snapshots.StatusCreated {
		t.Fatalf("unexpected status: %q", created.Status)
	}
	if created.SnapshotID == "" {
		t.Fatalf("missing snapshotId")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/snapshots/"+created.SnapshotID, nil)
	getRes := httptest.NewRecorder()
	router.ServeHTTP(getRes, getReq)
	if getRes.Code != http.StatusOK {
		t.Fatalf("unexpected get status: %d body=%s", getRes.Code, getRes.Body.String())
	}

	var fetched snapshots.Snapshot
	if err := json.Unmarshal(getRes.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("unmarshal get response: %v body=%s", err, getRes.Body.String())
	}
	if fetched.SnapshotID != created.SnapshotID {
		t.Fatalf("snapshotId mismatch: created=%q fetched=%q", created.SnapshotID, fetched.SnapshotID)
	}
}

func TestGetSnapshot_NotFound(t *testing.T) {
	t.Parallel()

	store := snapshots.NewStore()
	router := NewRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/snapshots/snap_missing", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}
}

func TestCreateSnapshot_InvalidMode(t *testing.T) {
	t.Parallel()

	store := snapshots.NewStore()
	router := NewRouter(store)

	createReq := map[string]any{
		"mode":      "Z",
		"sourceRef": map[string]any{"zip": map[string]any{"filename": "repo.zip"}},
	}
	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p123/snapshots", bytes.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}
}

func TestCreateSnapshot_MissingSourceRef(t *testing.T) {
	t.Parallel()

	store := snapshots.NewStore()
	router := NewRouter(store)

	createReq := map[string]any{
		"mode": "C",
	}
	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p123/snapshots", bytes.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}
}
