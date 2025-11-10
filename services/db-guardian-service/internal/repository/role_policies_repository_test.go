package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestRolePoliciesRepository_GetByProject_Positive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	r := NewRolePoliciesRepository(db)
	ctx := context.Background()

rows := sqlmock.NewRows([]string{"id","project_id","spec_yaml","version","created_at","updated_at"}).
AddRow("pol-1","p1","spec: x",1, time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, spec_yaml, version, created_at, updated_at FROM role_policies WHERE project_id = $1")).
		WithArgs("p1").
		WillReturnRows(rows)

	p, err := r.GetByProject(ctx, "p1")
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if p.ID != "pol-1" || p.ProjectID != "p1" { t.Fatalf("bad policy") }
}

func TestRolePoliciesRepository_Upsert_Positive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	r := NewRolePoliciesRepository(db)
	ctx := context.Background()
	p := &models.RolePolicy{ProjectID: "p1", SpecYAML: "spec: x", Version: 1}

	rows := sqlmock.NewRows([]string{"id","version"}).AddRow("pol-1", 2)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO role_policies (project_id, spec_yaml, version) VALUES ($1, $2, COALESCE($3,1)) ON CONFLICT (project_id) DO UPDATE SET spec_yaml = EXCLUDED.spec_yaml, version = role_policies.version + 1, updated_at = NOW() RETURNING id, version")).
		WithArgs("p1","spec: x",1).
		WillReturnRows(rows)

	id, err := r.Upsert(ctx, p)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if id != "pol-1" || p.Version != 2 { t.Fatalf("bad upsert") }
}
