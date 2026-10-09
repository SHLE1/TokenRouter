package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestFindResolvesIDsOnce 检查同一详情查询共用候选 ID，并保留各来源的权限参数。
func TestFindResolvesIDsOnce(t *testing.T) {
	for _, admin := range []bool{false, true} {
		name := "user"
		if admin {
			name = "admin"
		}
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = db.Close() }()
			ids := pq.Array([]string{"local-request", "client:legacy"})
			mock.ExpectQuery(`SELECT ARRAY\(SELECT id FROM request_lookup_ids\(\$1\)\)`).
				WithArgs("external-id").WillReturnRows(sqlmock.NewRows([]string{"ids"}).AddRow(`{"local-request","client:legacy"}`))
			mock.ExpectQuery(`SELECT record FROM request_records`).WithArgs(ids, int64(42), admin).
				WillReturnRows(sqlmock.NewRows([]string{"record"}))
			mock.ExpectQuery(`SELECT .* FROM usage_logs`).WithArgs(ids, int64(42), admin, "external-id").
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectQuery(`SELECT .* FROM ops_error_logs`).WithArgs(ids, int64(42), admin).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			if admin {
				mock.ExpectQuery(`SELECT .* FROM audit_logs`).WithArgs(ids).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}
			items, err := NewStore(db).Find(t.Context(), "external-id", 42, admin)
			require.NoError(t, err)
			require.Empty(t, items)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestFindStopsWhenLookupIsCanceled ID 解析被取消时结束查询，后续来源不继续读库。
func TestFindStopsWhenLookupIsCanceled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectQuery(`SELECT ARRAY`).WithArgs("id").WillReturnError(context.Canceled)
	_, err = NewStore(db).Find(t.Context(), "id", 42, false)
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, mock.ExpectationsWereMet())
}
