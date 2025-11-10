package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

func TestSLORepository_Create_And_GetByID(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewSLORepository(pdb)

	slo := &models.SLO{
		ID:          "11111111-1111-1111-1111-111111111111",
		ProjectID:   "proj-1",
		ServiceName: "svc",
		Type:        models.SLOTypeAvailability,
		Target:      99.9,
		Window:      24 * time.Hour,
		Query:       "promql",
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO slos (id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,'active',NOW(),NOW())")).
		WithArgs(slo.ID, slo.ProjectID, slo.ServiceName, string(slo.Type), slo.Target, int64(slo.Window.Seconds()), slo.Query).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), slo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "project_id", "service_name", "type", "target", "window_seconds", "query", "status", "created_at", "updated_at"}).
		AddRow(slo.ID, slo.ProjectID, slo.ServiceName, string(slo.Type), slo.Target, int64(slo.Window.Seconds()), slo.Query, "active", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at FROM slos WHERE id = $1")).
		WithArgs(slo.ID).WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), slo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != slo.ID {
		t.Errorf("want id %s, got %s", slo.ID, got.ID)
	}
}

func TestSLORepository_SaveAndGetStatus(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewSLORepository(pdb)

	status := &models.SLOStatus{SLOID: "22222222-2222-2222-2222-222222222222", Compliance: 99.0, BurnRate: 0.1, RemainingBudget: 1.0, LastCalculated: time.Now()}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO slo_status (slo_id, compliance, burn_rate, remaining_budget, last_calculated) VALUES ($1,$2,$3,$4,NOW()) ON CONFLICT (slo_id) DO UPDATE SET compliance = EXCLUDED.compliance, burn_rate = EXCLUDED.burn_rate, remaining_budget = EXCLUDED.remaining_budget, last_calculated = NOW()")).
		WithArgs(status.SLOID, status.Compliance, status.BurnRate, status.RemainingBudget).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.SaveStatus(context.Background(), status); err != nil {
		t.Fatalf("SaveStatus: %v", err)
	}

	rows := sqlmock.NewRows([]string{"slo_id", "compliance", "burn_rate", "remaining_budget", "last_calculated"}).
		AddRow(status.SLOID, status.Compliance, status.BurnRate, status.RemainingBudget, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT slo_id, compliance, burn_rate, remaining_budget, last_calculated FROM slo_status WHERE slo_id = $1")).
		WithArgs(status.SLOID).WillReturnRows(rows)

	got, err := repo.GetStatus(context.Background(), status.SLOID)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.SLOID != status.SLOID {
		t.Errorf("want slo_id %s, got %s", status.SLOID, got.SLOID)
	}
}

func TestSLORepository_GetHistory(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewSLORepository(pdb)

	from := time.Now().Add(-time.Hour)
	to := time.Now()

	rows := sqlmock.NewRows([]string{"slo_id", "timestamp", "compliance", "burn_rate"}).
		AddRow("33333333-3333-3333-3333-333333333333", time.Now(), 99.5, 0.05)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT slo_id, timestamp, compliance, burn_rate FROM slo_history WHERE slo_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp DESC LIMIT 1000")).
		WithArgs("33333333-3333-3333-3333-333333333333", from, to).WillReturnRows(rows)

	h, err := repo.GetHistory(context.Background(), "33333333-3333-3333-3333-333333333333", from, to)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(h) != 1 {
		t.Fatalf("want 1 history row, got %d", len(h))
	}
}

func TestSLORepository_ListAll_ReturnsRows(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewSLORepository(pdb)

	rows := sqlmock.NewRows([]string{"id","project_id","service_name","type","target","window_seconds","query","status","created_at","updated_at"}).
		AddRow("111", "p1", "svc1", string(models.SLOTypeAvailability), 99.9, int64(300), "q1", "active", time.Now(), time.Now()).
		AddRow("222", "p2", "svc2", string(models.SLOTypeLatency), 0.5, int64(600), "q2", "active", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at FROM slos WHERE status='active' ORDER BY created_at DESC")).
		WillReturnRows(rows)

	list, err := repo.ListAll(context.Background())
	if err != nil { t.Fatalf("ListAll: %v", err) }
	if len(list) != 2 { t.Fatalf("want 2, got %d", len(list)) }
	if list[0].ID != "111" || list[1].ID != "222" { t.Fatalf("unexpected IDs: %+v", []string{list[0].ID, list[1].ID}) }
	if list[0].Window != 5*time.Minute { t.Fatalf("window not parsed: %v", list[0].Window) }
}
