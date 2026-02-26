package snapshots

import "time"

type Mode string

const (
	ModeA Mode = "A"
	ModeB Mode = "B"
	ModeC Mode = "C"
)

type Status string

const (
	StatusCreated    Status = "created"
	StatusProcessing Status = "processing"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
)

type Snapshot struct {
	SnapshotID string           `json:"snapshotId"`
	ProjectID  string           `json:"projectId"`
	Mode       Mode             `json:"mode"`
	Status     Status           `json:"status"`
	CreatedAt  time.Time        `json:"createdAt"`
	SourceRef  map[string]any   `json:"sourceRef"`
	Artifacts  []map[string]any `json:"artifacts,omitempty"`
	Meta       map[string]any   `json:"meta,omitempty"`
}
