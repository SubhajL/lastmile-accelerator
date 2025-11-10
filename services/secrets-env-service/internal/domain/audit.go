package domain

import "time"

// AuditLogEntry represents an immutable audit record for secret metadata actions.
type AuditLogEntry struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ProjectID   string    `json:"project_id"`
	Key         string    `json:"key"`
	Environment string    `json:"environment"`
	Action      string    `json:"action"` // created|updated|deleted|accessed
	Actor       string    `json:"actor"`
	OccurredAt  time.Time `json:"occurred_at"`
	Metadata    map[string]any `json:"metadata"`
}
