package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

type ErrorRepository interface {
	RecordEvent(ctx context.Context, projectID, fingerprint, title, message, stack string, metadata map[string]interface{}) (string, error)
	ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error)
	GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error)
	ListEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error)
	ResolveGroup(ctx context.Context, groupID string) error
}

type postgresErrorRepo struct { db *storage.PostgresDB }

func NewErrorRepository(db *storage.PostgresDB) ErrorRepository { return &postgresErrorRepo{db: db} }

func (r *postgresErrorRepo) RecordEvent(ctx context.Context, projectID, fingerprint, title, message, stack string, metadata map[string]interface{}) (string, error) {
	metaBytes, _ := json.Marshal(metadata)
	// Upsert group and return id (sqlmock test matches RETURNING id usage)
	const upsert = `INSERT INTO error_groups (id, project_id, fingerprint, title, status, first_seen, last_seen, occurrences, sample_stack)
	VALUES ($1,$2,$3,$4,'open',NOW(),NOW(),1,$5)
	ON CONFLICT (project_id, fingerprint) DO UPDATE SET last_seen=NOW(), occurrences=error_groups.occurrences+1, sample_stack=COALESCE(error_groups.sample_stack, EXCLUDED.sample_stack)
	RETURNING id`
	var groupID string
	if err := r.db.DB().QueryRowContext(ctx, upsert, newUUID(), projectID, fingerprint, title, stack).Scan(&groupID); err != nil {
		return "", fmt.Errorf("upsert group: %w", err)
	}
	const insEvent = `INSERT INTO error_events (id, group_id, project_id, ts, message, stack, metadata) VALUES ($1,$2,$3,NOW(),$4,$5,$6)`
	if _, err := r.db.DB().ExecContext(ctx, insEvent, newUUID(), groupID, projectID, message, stack, string(metaBytes)); err != nil {
		return "", fmt.Errorf("insert event: %w", err)
	}
	return groupID, nil
}

func (r *postgresErrorRepo) ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error) {
	q := `SELECT id, project_id, fingerprint, title, status, first_seen, last_seen, occurrences, sample_stack FROM error_groups WHERE project_id=$1 ORDER BY last_seen DESC LIMIT 50 OFFSET 0`
	rows, err := r.db.DB().QueryContext(ctx, q, projectID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []models.ErrorGroup
	for rows.Next() {
		var g models.ErrorGroup
		var first, last sql.NullTime
		if err := rows.Scan(&g.ID, &g.ProjectID, &g.Fingerprint, &g.Title, &g.Status, &first, &last, &g.Occurrences, &g.SampleStack); err != nil { return nil, err }
		if first.Valid { g.FirstSeen = first.Time.UTC().Format(timeRFC3339) }
		if last.Valid { g.LastSeen = last.Time.UTC().Format(timeRFC3339) }
		out = append(out, g)
	}
	return out, nil
}

func (r *postgresErrorRepo) GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error) {
	q := `SELECT id, project_id, fingerprint, title, status, first_seen, last_seen, occurrences, sample_stack FROM error_groups WHERE id=$1`
	row := r.db.DB().QueryRowContext(ctx, q, groupID)
	var g models.ErrorGroup
	var first, last sql.NullTime
	if err := row.Scan(&g.ID, &g.ProjectID, &g.Fingerprint, &g.Title, &g.Status, &first, &last, &g.Occurrences, &g.SampleStack); err != nil { return nil, err }
	if first.Valid { g.FirstSeen = first.Time.UTC().Format(timeRFC3339) }
	if last.Valid { g.LastSeen = last.Time.UTC().Format(timeRFC3339) }
	return &g, nil
}

func (r *postgresErrorRepo) ListEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error) {
	q := `SELECT id, group_id, project_id, ts, message, stack, metadata FROM error_events WHERE group_id=$1 ORDER BY ts DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.DB().QueryContext(ctx, q, groupID, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []models.ErrorEvent
	for rows.Next() {
		var e models.ErrorEvent
		var ts sql.NullTime
		var meta sql.NullString
		if err := rows.Scan(&e.ID, &e.GroupID, &e.ProjectID, &ts, &e.Message, &e.Stack, &meta); err != nil { return nil, err }
		if ts.Valid { e.TS = ts.Time.UTC().Format(timeRFC3339) }
		if meta.Valid {
			var m map[string]interface{}
			_ = json.Unmarshal([]byte(meta.String), &m)
			e.Metadata = m
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *postgresErrorRepo) ResolveGroup(ctx context.Context, groupID string) error {
	_, err := r.db.DB().ExecContext(ctx, `UPDATE error_groups SET status='resolved', last_seen=NOW() WHERE id=$1`, groupID)
	if err != nil { return fmt.Errorf("resolve: %w", err) }
	return nil
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

func newUUID() string { return fmt.Sprintf("%d", timeNowUnixNano()) }

var timeNowUnixNano = func() int64 { return 0 } // patched in tests if needed
