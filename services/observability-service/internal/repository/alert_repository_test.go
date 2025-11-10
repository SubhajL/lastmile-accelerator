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

func TestAlertRepository_CreateAndGetByID(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewAlertRepository(pdb)

	rule := &models.AlertRule{
		ID:        "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		SLOID:     "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		Threshold: 95.0,
		Channels:  []models.AlertChannel{models.AlertChannelEmail},
		Enabled:   true,
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO alert_rules (id, slo_id, threshold, channels, enabled, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,NOW(),NOW())")).
		WithArgs(rule.ID, rule.SLOID, rule.Threshold, sqlmock.AnyArg(), rule.Enabled).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), rule); err != nil {
		t.Fatalf("Create: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "slo_id", "threshold", "channels", "enabled", "created_at", "updated_at"}).
		AddRow(rule.ID, rule.SLOID, rule.Threshold, `{"channels":["email"]}`, true, time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slo_id, threshold, channels, enabled, created_at, updated_at FROM alert_rules WHERE id=$1")).
		WithArgs(rule.ID).WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != rule.ID {
		t.Errorf("want id %s got %s", rule.ID, got.ID)
	}
}

func TestAlertRepository_GetBySLO_AndHistory(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewAlertRepository(pdb)

	sloID := "cccccccc-cccc-cccc-cccc-cccccccccccc"

	rows := sqlmock.NewRows([]string{"id", "slo_id", "threshold", "channels", "enabled", "created_at", "updated_at"}).
		AddRow("id1", sloID, 90.0, `{"channels":["slack"]}`, true, time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, slo_id, threshold, channels, enabled, created_at, updated_at FROM alert_rules WHERE slo_id=$1")).
		WithArgs(sloID).WillReturnRows(rows)

	rules, err := repo.GetBySLO(context.Background(), sloID)
	if err != nil {
		t.Fatalf("GetBySLO: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("want 1, got %d", len(rules))
	}

	hRow := sqlmock.NewRows([]string{"id", "alert_rule_id", "slo_id", "timestamp", "compliance", "threshold", "notified"}).
		AddRow("hid1", "id1", sloID, time.Now(), 85.0, 90.0, true)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, alert_rule_id, slo_id, timestamp, compliance, threshold, notified FROM alert_history WHERE alert_rule_id=$1 ORDER BY timestamp DESC LIMIT 100")).
		WithArgs("id1").WillReturnRows(hRow)

	h, err := repo.GetHistory(context.Background(), "id1", 100)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(h) != 1 {
		t.Fatalf("want 1, got %d", len(h))
	}
}
