//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/pkg/timezone"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/settings/preaggregation"
	"github.com/TokenFlux/TokenRouter/internal/usage"
	"github.com/TokenFlux/TokenRouter/internal/usage/postgres/query"
	"github.com/TokenFlux/TokenRouter/migrations"
	"github.com/stretchr/testify/require"
)

// TestUserActivityTransactions 对照原始记录检查回退、归属变更、去重和回滚。
func TestUserActivityTransactions(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()
	u := mustCreateUser(t, client, &identity.User{Email: "activity@test.local"})
	v := mustCreateUser(t, client, &identity.User{Email: "activity-other@test.local"})
	k := mustCreateApiKey(t, client, &apikey.APIKey{UserID: u.ID, Key: "activity-key"})
	p := mustCreateProvider(t, client, &provider.Record{Name: "activity"})
	insert := `INSERT INTO usage_logs(user_id,api_key_id,provider_id,model,request_id,created_at)
		VALUES($1,$2,$3,'test',$4,$5) ON CONFLICT(request_id,api_key_id) DO NOTHING`
	old := time.Now().UTC().Truncate(time.Second).Add(-24 * time.Hour)
	latest := old.Add(time.Hour)
	for i, at := range []time.Time{latest, old, latest} {
		_, err := integrationDB.ExecContext(ctx, insert, u.ID, k.ID, p.ID, fmt.Sprint(i), at)
		require.NoError(t, err)
	}
	assertActivityMatchesLogs(t, integrationDB)
	got, err := query.GetLatestUsedAtByUserIDs(ctx, integrationDB, []int64{u.ID, v.ID, u.ID})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.True(t, latest.Equal(*got[u.ID]))
	_, err = integrationDB.ExecContext(ctx, insert, u.ID, k.ID, p.ID, "0", latest.Add(time.Hour))
	require.NoError(t, err)
	assertActivityMatchesLogs(t, integrationDB)
	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, insert, u.ID, k.ID, p.ID, "rollback", latest.Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	_, err = integrationDB.ExecContext(ctx, "DELETE FROM usage_logs WHERE user_id=$1 AND created_at=$2", u.ID, latest)
	require.NoError(t, err)
	assertActivityMatchesLogs(t, integrationDB)
	_, err = integrationDB.ExecContext(ctx, "UPDATE usage_logs SET user_id=$1,created_at=$2 WHERE user_id=$3", v.ID, latest, u.ID)
	require.NoError(t, err)
	assertActivityMatchesLogs(t, integrationDB)
	// 在已有数据上重复执行迁移，回填结果也需与明细一致。
	migration, err := migrations.FS.ReadFile("289_usage_user_activity.sql")
	require.NoError(t, err)
	for range 2 {
		_, err = integrationDB.ExecContext(ctx, string(migration))
		require.NoError(t, err)
		assertActivityMatchesLogs(t, integrationDB)
	}
	_, err = integrationDB.ExecContext(ctx, "TRUNCATE usage_logs CASCADE")
	require.NoError(t, err)
	assertActivityMatchesLogs(t, integrationDB)
}

// assertActivityMatchesLogs 比较全部用户，缺行和多余汇总行都会失败。
func assertActivityMatchesLogs(t *testing.T, db *sql.DB) {
	t.Helper()
	var differences int
	err := db.QueryRow(`WITH expected AS (SELECT user_id,MAX(created_at) AS last_used_at FROM usage_logs GROUP BY user_id)
		SELECT COUNT(*) FROM (
		 (SELECT * FROM expected EXCEPT SELECT * FROM usage_user_activity)
		 UNION ALL (SELECT * FROM usage_user_activity EXCEPT SELECT * FROM expected)
		) differences`).Scan(&differences)
	require.NoError(t, err)
	require.Zero(t, differences)
}

// TestUserActivityPartitionDrop 使用隔离 schema 检查分区删除、回滚和目标校验。
func TestUserActivityPartitionDrop(t *testing.T) {
	ctx := context.Background()
	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(`CREATE SCHEMA activity_partition_test;
		SET LOCAL search_path=activity_partition_test,public;
		CREATE TABLE users(id bigint PRIMARY KEY);
		CREATE TABLE usage_logs(user_id bigint,created_at timestamptz NOT NULL) PARTITION BY RANGE(created_at);
		CREATE INDEX ON usage_logs(user_id,created_at);
		CREATE TABLE usage_logs_202601 PARTITION OF usage_logs FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
		CREATE TABLE usage_logs_202602 PARTITION OF usage_logs FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
		INSERT INTO users VALUES(1),(2)`)
	require.NoError(t, err)
	migration, err := migrations.FS.ReadFile("289_usage_user_activity.sql")
	require.NoError(t, err)
	_, err = tx.Exec(string(migration))
	require.NoError(t, err)
	_, err = tx.Exec(`INSERT INTO usage_logs VALUES(1,'2026-01-10'),(1,'2026-02-10'),(2,'2026-01-20'); SAVEPOINT before_drop`)
	require.NoError(t, err)
	repo := NewAggregationStoreWithSQL(tx, timezone.NewCalendar(time.UTC))
	require.NoError(t, repo.CleanupUsageLogs(ctx, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)))
	var count int
	require.NoError(t, tx.QueryRow("SELECT COUNT(*) FROM usage_user_activity").Scan(&count))
	require.Equal(t, 1, count)
	_, err = tx.Exec("ROLLBACK TO SAVEPOINT before_drop")
	require.NoError(t, err)
	require.NoError(t, tx.QueryRow("SELECT COUNT(*) FROM usage_user_activity").Scan(&count))
	require.Equal(t, 2, count)
	_, err = tx.Exec("SELECT usage_drop_partition_with_activity('users')")
	require.ErrorContains(t, err, "直属分区")
}

// TestUserActivityConcurrentWrites 覆盖不同连接对同用户的批量插入和清理交错。
func TestUserActivityConcurrentWrites(t *testing.T) {
	client := testEntClient(t)
	u := mustCreateUser(t, client, &identity.User{Email: "activity-concurrent@test.local"})
	k := mustCreateApiKey(t, client, &apikey.APIKey{UserID: u.ID, Key: "activity-concurrent"})
	p := mustCreateProvider(t, client, &provider.Record{Name: "activity-concurrent"})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	start := make(chan struct{})
	errors := make(chan error, 8)
	var workers sync.WaitGroup
	for worker := range 8 {
		workers.Go(func() {
			<-start
			for iteration := range 10 {
				request := fmt.Sprintf("activity-%d-%d", worker, iteration)
				_, err := integrationDB.ExecContext(ctx, `INSERT INTO usage_logs(user_id,api_key_id,provider_id,model,request_id,created_at)
					SELECT $1,$2,$3,'test',$4||n::text,now()+n*interval '1 second' FROM generate_series(1,3) n`, u.ID, k.ID, p.ID, request)
				if err == nil {
					_, err = integrationDB.ExecContext(ctx, "DELETE FROM usage_logs WHERE request_id LIKE $1 AND api_key_id=$2", request+"%", k.ID)
				}
				if err != nil {
					errors <- err
					return
				}
			}
		})
	}
	close(start)
	workers.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	assertActivityMatchesLogs(t, integrationDB)
}

// TestAPIKeyDashboardAnalyticsParity 包含时区跨日、空耗时、零费用及未来记录。
func (s *UsageLogRepoSuite) TestAPIKeyDashboardAnalyticsParity() {
	user := mustCreateUser(s.T(), s.client, &identity.User{Email: "key-analytics@test.local"})
	key := mustCreateApiKey(s.T(), s.client, &apikey.APIKey{UserID: user.ID, Key: "key-analytics"})
	p := mustCreateProvider(s.T(), s.client, &provider.Record{Name: "key-analytics"})
	loc, err := time.LoadLocation("Asia/Kathmandu")
	s.Require().NoError(err)
	s.repo.calendar = timezone.NewCalendar(loc)
	now := time.Now().UTC().Truncate(time.Second)
	for i, at := range []time.Time{now.Add(-120 * time.Hour), now.Add(-53 * time.Hour), now.Add(-12 * time.Hour), now.Add(-2 * time.Hour), now.Add(48 * time.Hour)} {
		log := s.createUsageLog(user, key, p, 10+i, 5+i, float64(i), at)
		if i%2 == 0 {
			_, err = s.tx.ExecContext(s.ctx, "UPDATE usage_logs SET duration_ms=$1 WHERE id=$2", i*100, log.ID)
			s.Require().NoError(err)
		}
	}
	want, err := s.repo.GetAPIKeyDashboardStats(s.ctx, key.ID)
	s.Require().NoError(err)
	start := now.Add(-72 * time.Hour).Truncate(time.Hour)
	watermark := now.Add(-3 * time.Hour)
	aggregation := NewAggregationStoreWithSQL(s.tx, s.repo.calendar)
	s.Require().NoError(aggregation.AggregateUsageAnalyticsRange(s.ctx, start, watermark))
	s.Require().NoError(aggregation.SaveUsageAnalyticsAggregationState(s.ctx, &usage.UsageAnalyticsAggregationState{
		LiveWatermark: watermark, CoverageStart: &start, Phase: "idle",
	}))
	s.repo.preAggregation = preaggregation.NewPreAggregationSettingsService(nil, &preaggregation.Options{
		Usage: preaggregation.UsageOptions{Enabled: true, IntervalSeconds: 60},
	})
	_, ok, err := s.repo.getAPIKeyDashboardStatsFromAnalytics(s.ctx, key.ID)
	s.Require().NoError(err)
	s.Require().True(ok)
	got, err := s.repo.GetAPIKeyDashboardStats(s.ctx, key.ID)
	s.Require().NoError(err)
	s.Require().Equal(want, got)
	// 删除最早记录后，累计范围应排除此前仍留在聚合表中的历史。
	_, err = s.tx.ExecContext(s.ctx, "DELETE FROM usage_logs WHERE api_key_id=$1 AND created_at<$2", key.ID, now.Add(-24*time.Hour))
	s.Require().NoError(err)
	got, err = s.repo.GetAPIKeyDashboardStats(s.ctx, key.ID)
	s.Require().NoError(err)
	s.repo.preAggregation = nil
	want, err = s.repo.GetAPIKeyDashboardStats(s.ctx, key.ID)
	s.Require().NoError(err)
	s.Require().Equal(want, got)
}
