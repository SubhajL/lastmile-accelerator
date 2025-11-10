package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"example.com/lma/secrets-env-service/internal/domain"
)

type AuditLogRepositoryPG struct{ db *sql.DB }

func NewAuditLogRepositoryPostgres(db *sql.DB) *AuditLogRepositoryPG { return &AuditLogRepositoryPG{db: db} }

func (r *AuditLogRepositoryPG) Write(ctx context.Context, e *domain.AuditLogEntry) error {
	var meta []byte
	if e.Metadata != nil {
		b, _ := json.Marshal(e.Metadata)
		meta = b
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO audit_logs (tenant_id, project_id, secret_key, action, actor, timestamp, metadata) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.TenantID, e.ProjectID, e.Key, e.Action, e.Actor, e.OccurredAt, meta,
	)
	return err
}
