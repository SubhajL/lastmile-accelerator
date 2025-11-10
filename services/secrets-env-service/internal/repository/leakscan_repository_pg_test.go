package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestLeakScanRepositoryPG_CreateBatch_MultipleRows(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := NewLeakScanRepositoryPostgres(db)

	scans := []*domain.ClientLeakScan{
		{ID:"1", SnapshotID:"s", FilePath:"a.js", LineNumber:1, Pattern:"aws", Severity:"high", CreatedAt: time.Unix(1,0)},
		{ID:"2", SnapshotID:"s", FilePath:"b.js", LineNumber:2, Pattern:"jwt", Severity:"low", CreatedAt: time.Unix(2,0)},
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO client_leak_scans (id, snapshot_id, file_path, line_number, pattern, severity, fixed, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8),($9,$10,$11,$12,$13,$14,$15,$16)`)).
		WithArgs(
			"1","s","a.js",1,"aws","high",false, time.Unix(1,0),
			"2","s","b.js",2,"jwt","low", false, time.Unix(2,0),
		).
		WillReturnResult(sqlmock.NewResult(0,2))

	require.NoError(t, r.CreateBatch(context.Background(), scans))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLeakScanRepositoryPG_GetBySnapshot_FilterSeverity(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := NewLeakScanRepositoryPostgres(db)
	rows := sqlmock.NewRows([]string{"id","snapshot_id","file_path","line_number","pattern","severity","fixed","created_at"}).
		AddRow("1","s","a.js",1,"aws","high",false, time.Unix(1,0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, snapshot_id, file_path, line_number, pattern, severity, fixed, created_at FROM client_leak_scans WHERE snapshot_id=$1 AND ($2='' OR severity=$2) ORDER BY created_at ASC`)).
		WithArgs("s","high").
		WillReturnRows(rows)
	res, err := r.GetBySnapshotID(context.Background(), "s", "high")
	require.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestLeakScanRepositoryPG_MarkAsFixed(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := NewLeakScanRepositoryPostgres(db)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE client_leak_scans SET fixed=true, fixed_at=NOW() WHERE id=$1`)).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0,1))
	require.NoError(t, r.MarkAsFixed(context.Background(), "1"))
	mock.ExpectationsWereMet()
}
