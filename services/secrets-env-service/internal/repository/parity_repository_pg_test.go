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


func TestParityRepositoryPG_Create_InsertsJSONB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	r := NewParityRepositoryPostgres(db)

	check := &domain.EnvParityCheck{
		ProjectID:     "p",
		ScanTimestamp: time.Unix(1700000000, 0),
		MissingKeys:   []string{"A","B"},
		ExtraKeys:     []string{"C"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO env_parity_checks (project_id, scan_timestamp, missing_keys, extra_keys, has_drift) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs("p", check.ScanTimestamp, sqlmock.AnyArg(), sqlmock.AnyArg(), true).
		WillReturnResult(sqlmock.NewResult(1,1))

	require.NoError(t, r.Create(ctx(), check))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestParityRepositoryPG_GetLatest_ReturnsNewest(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	r := NewParityRepositoryPostgres(db)

	rows := sqlmock.NewRows([]string{"project_id","scan_timestamp","missing_keys","extra_keys","has_drift"}).
		AddRow("p", time.Unix(2,0), `{"A"}`, `[]`, true)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT project_id, scan_timestamp, missing_keys, extra_keys, has_drift FROM env_parity_checks WHERE project_id=$1 ORDER BY scan_timestamp DESC LIMIT 1`)).
		WithArgs("p").
		WillReturnRows(rows)

	res, err := r.GetLatest(ctx(), "p")
	require.NoError(t, err)
	assert.Equal(t, "p", res.ProjectID)
}

func TestParityRepositoryPG_GetHistory_LimitsResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	r := NewParityRepositoryPostgres(db)

	rows := sqlmock.NewRows([]string{"project_id","scan_timestamp","missing_keys","extra_keys","has_drift"}).
		AddRow("p", time.Unix(3,0), `[]`, `[]`, false).
		AddRow("p", time.Unix(2,0), `[]`, `[]`, false)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT project_id, scan_timestamp, missing_keys, extra_keys, has_drift FROM env_parity_checks WHERE project_id=$1 ORDER BY scan_timestamp DESC LIMIT $2`)).
		WithArgs("p", 2).
		WillReturnRows(rows)

	h, err := r.GetHistory(ctx(), "p", 2)
	require.NoError(t, err)
	assert.Len(t, h, 2)
}

func ctx() context.Context { return context.Background() }
