package repository

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/DATA-DOG/go-sqlmock"
)

type anyTime struct{}
func (a anyTime) Match(v driver.Value) bool { _, ok := v.(time.Time); return ok }

func TestAuditLogRepositoryPG_Write_InsertsRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	r := NewAuditLogRepositoryPostgres(db)

	e := &domain.AuditLogEntry{TenantID:"t", ProjectID:"p", Key:"K", Action:"created", Actor:"u", OccurredAt: time.Now(), Metadata: map[string]any{"x":"y"}}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO audit_logs")).
		WithArgs("t","p","K","created","u", anyTime{}, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1,1))

	if err := r.Write(context.Background(), e); err != nil { t.Fatalf("write err: %v", err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("unmet: %v", err) }
}
