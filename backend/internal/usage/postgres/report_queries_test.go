package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/TokenFlux/TokenRouter/internal/settings/preaggregation"
	"github.com/TokenFlux/TokenRouter/internal/team"
	"github.com/stretchr/testify/require"
)

// TestProviderReportOwnsJITTransaction 将 JIT 设置限定在报表事务，错误路径也结束事务。
func TestProviderReportOwnsJITTransaction(t *testing.T) {
	for _, fail := range []bool{false, true} {
		db, mock := newSQLMock(t)
		r := &Store{sql: db, db: db}
		mock.ExpectBegin()
		mock.ExpectExec("SET LOCAL jit = off").WillReturnResult(sqlmock.NewResult(0, 0))
		query := mock.ExpectQuery("(?s)WITH scoped.*GROUP BY GROUPING SETS")
		if fail {
			query.WillReturnError(errors.New("report unavailable"))
			mock.ExpectRollback()
		} else {
			query.WillReturnRows(sqlmock.NewRows([]string{"kind", "label", "requests", "input", "output", "creation", "read", "cost", "user_cost", "provider_cost", "duration", "count"}))
			mock.ExpectCommit()
		}
		_, err := r.GetProviderUsageStats(context.Background(), 1, time.Now().Add(-time.Hour), time.Now())
		if fail {
			require.ErrorContains(t, err, "report unavailable")
		} else {
			require.NoError(t, err)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

// TestStatsEndpointSourceQueries 未选中的端点维度不向数据库发送查询。
func TestStatsEndpointSourceQueries(t *testing.T) {
	for _, source := range []string{"", "inbound", "upstream", "path"} {
		db, mock := newSQLMock(t)
		r := &Store{sql: db}
		mock.ExpectQuery("(?s)SELECT.*total_requests.*FROM usage_logs").WillReturnRows(sqlmock.NewRows([]string{"requests", "input", "output", "cache", "creation", "read", "cost", "actual", "provider", "average"}).AddRow(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
		for _, kind := range []string{"inbound", "upstream", "path"} {
			if source != "" && source != kind {
				continue
			}
			pattern := "(?s)SELECT.*TRIM\\(" + kind + "_endpoint\\).*GROUP BY endpoint"
			if kind == "path" {
				pattern = "(?s)SELECT.*CONCAT.*GROUP BY endpoint"
			}
			mock.ExpectQuery(pattern).WillReturnRows(sqlmock.NewRows([]string{"endpoint", "requests", "tokens", "cost", "actual"}).AddRow(kind, 1, 2, 3, 4))
		}
		result, err := r.GetStatsWithFilters(context.Background(), UsageLogFilters{EndpointSource: source})
		require.NoError(t, err)
		if source == "" || source == "inbound" {
			require.Len(t, result.Endpoints, 1)
		} else {
			require.Empty(t, result.Endpoints)
		}
		if source == "" || source == "upstream" {
			require.Len(t, result.UpstreamEndpoints, 1)
		} else {
			require.Empty(t, result.UpstreamEndpoints)
		}
		if source == "" || source == "path" {
			require.Len(t, result.EndpointPaths, 1)
		} else {
			require.Empty(t, result.EndpointPaths)
		}
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

// TestTeamReportAggregationFailureUsesRaw 聚合状态故障时完整读取原始范围。
func TestTeamReportAggregationFailureUsesRaw(t *testing.T) {
	db, mock := newSQLMock(t)
	r := &Store{sql: db, preAggregation: preaggregation.NewPreAggregationSettingsService(nil, &preaggregation.Options{Usage: preaggregation.UsageOptions{Enabled: true}})}
	mock.ExpectQuery("SELECT MIN\\(ul.created_at\\)").WillReturnRows(sqlmock.NewRows([]string{"oldest"}).AddRow(time.Now().Add(-30 * time.Minute)))
	mock.ExpectQuery("SELECT live_watermark, coverage_start").WillReturnError(errors.New("state unavailable"))
	mock.ExpectQuery("(?s)WITH scoped.*FROM usage_logs.*GROUP BY GROUPING SETS").WillReturnRows(sqlmock.NewRows([]string{"day", "cost", "requests", "input", "output"}).AddRow(nil, 12, 3, 4, 5))
	result, err := r.GetTeamUsageSummary(context.Background(), 7, team.TeamUsageQuery{From: time.Now().Add(-time.Hour), To: time.Now()})
	require.NoError(t, err)
	require.EqualValues(t, 12, result.ActualCost)
	require.EqualValues(t, 3, result.RequestCount)
	require.NoError(t, mock.ExpectationsWereMet())
}
