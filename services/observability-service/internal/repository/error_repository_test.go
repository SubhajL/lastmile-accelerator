package repository

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

type eventIn struct {
	ProjectID string
	Message   string
	Stack     string
	Fingerprint string
	Title     string
	Metadata  map[string]interface{}
}

func TestErrorRepository_RecordEvent_CreatesGroupAndEvent(t *testing.T) {
	mdb, mock, err := sqlmock.New(); if err != nil { t.Fatal(err) }
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewErrorRepository(pdb)

	e := eventIn{ ProjectID:"p1", Message:"boom", Stack:"at a()", Fingerprint:"abc", Title:"boom", Metadata: map[string]interface{}{"k":"v"} }
	meta, _ := json.Marshal(e.Metadata)

	// Upsert group
mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO error_groups (id, project_id, fingerprint, title, status, first_seen, last_seen, occurrences, sample_stack) VALUES ($1,$2,$3,$4,'open',NOW(),NOW(),1,$5) ON CONFLICT (project_id, fingerprint) DO UPDATE SET last_seen=NOW(), occurrences=error_groups.occurrences+1, sample_stack=COALESCE(error_groups.sample_stack, EXCLUDED.sample_stack) RETURNING id")).
		WithArgs(sqlmock.AnyArg(), e.ProjectID, e.Fingerprint, e.Title, e.Stack).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("g1"))

	// Insert event
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO error_events (id, group_id, project_id, ts, message, stack, metadata) VALUES ($1,$2,$3,NOW(),$4,$5,$6)")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), e.ProjectID, e.Message, e.Stack, string(meta)).
		WillReturnResult(sqlmock.NewResult(1,1))

	gid, err := repo.RecordEvent(context.Background(), e.ProjectID, e.Fingerprint, e.Title, e.Message, e.Stack, e.Metadata)
	if err != nil { t.Fatalf("RecordEvent: %v", err) }
	if gid == "" { t.Fatalf("expected group id") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("unmet: %v", err) }
}

func TestErrorRepository_ListGroups_Filters(t *testing.T) {
	mdb, mock, err := sqlmock.New(); if err != nil { t.Fatal(err) }
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewErrorRepository(pdb)

	rows := sqlmock.NewRows([]string{"id","project_id","fingerprint","title","status","first_seen","last_seen","occurrences","sample_stack"}).
		AddRow("g1","p1","f1","Boom","open", time.Now(), time.Now(), 2, "stack")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, fingerprint, title, status, first_seen, last_seen, occurrences, sample_stack FROM error_groups WHERE project_id=$1 ORDER BY last_seen DESC LIMIT 50 OFFSET 0")).
		WithArgs("p1").WillReturnRows(rows)

	gs, err := repo.ListGroups(context.Background(), "p1", models.ErrorGroupFilter{})
	if err != nil || len(gs) != 1 { t.Fatalf("ListGroups: %v, len=%d", err, len(gs)) }
}

func TestErrorRepository_ResolveGroup(t *testing.T) {
	mdb, mock, err := sqlmock.New(); if err != nil { t.Fatal(err) }
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewErrorRepository(pdb)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE error_groups SET status='resolved', last_seen=NOW() WHERE id=$1")).WithArgs("g1").WillReturnResult(sqlmock.NewResult(1,1))
	if err := repo.ResolveGroup(context.Background(), "g1"); err != nil { t.Fatalf("ResolveGroup: %v", err) }
}
