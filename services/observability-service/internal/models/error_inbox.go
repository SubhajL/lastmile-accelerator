package models

type ErrorGroup struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	Fingerprint string   `json:"fingerprint"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	FirstSeen   string   `json:"first_seen"`
	LastSeen    string   `json:"last_seen"`
	Occurrences int64    `json:"occurrences"`
	SampleStack string   `json:"sample_stack"`
}

type ErrorEvent struct {
	ID        string                 `json:"id"`
	GroupID   string                 `json:"group_id"`
	ProjectID string                 `json:"project_id"`
	TS        string                 `json:"ts"`
	Message   string                 `json:"message"`
	Stack     string                 `json:"stack"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ErrorGroupFilter struct {
	Status string
	From   string
	To     string
	Query  string
}
