//go:build integration

package postgres

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/usage"
	"github.com/stretchr/testify/require"
)

// BenchmarkUserActivityWrites 使用完整迁移和生产批量写入 SQL 比较活动汇总的开销。
func BenchmarkUserActivityWrites(b *testing.B) {
	benchmarkUserActivityWrites(b, false)
}

// BenchmarkUserActivityConcurrentWrites 比较并发连接写入同用户及多个用户的吞吐。
func BenchmarkUserActivityConcurrentWrites(b *testing.B) {
	benchmarkUserActivityWrites(b, true)
}

// benchmarkUserActivityWrites 在每个子基准结束时恢复触发器并清理夹具。
func benchmarkUserActivityWrites(b *testing.B, concurrent bool) {
	ctx := context.Background()
	client := integrationEntClient
	for _, spread := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			b.Run(fmt.Sprintf("spread=%t/activity=%t", spread, enabled), func(b *testing.B) {
				mode := "DISABLE"
				if enabled {
					mode = "ENABLE"
				}
				_, err := integrationDB.Exec("ALTER TABLE usage_logs " + mode + " TRIGGER usage_activity_insert")
				require.NoError(b, err)
				b.Cleanup(func() {
					_, e := integrationDB.Exec("ALTER TABLE usage_logs ENABLE TRIGGER usage_activity_insert; TRUNCATE users CASCADE")
					require.NoError(b, e)
				})
				p, err := client.Provider.Create().SetName("activity-bench").SetPlatform("openai").SetType("apikey").SetCredentials(map[string]any{}).Save(ctx)
				require.NoError(b, err)
				userIDs, keyIDs := make([]int64, 64), make([]int64, 64)
				for i := range userIDs {
					if i > 0 && !spread {
						userIDs[i], keyIDs[i] = userIDs[0], keyIDs[0]
						continue
					}
					u, e := client.User.Create().SetEmail(fmt.Sprintf("%d@activity-bench.test", i)).SetPasswordHash("hash").Save(ctx)
					require.NoError(b, e)
					k, e := client.APIKey.Create().SetUserID(u.ID).SetKey(fmt.Sprintf("activity-bench-%d", i)).SetName("bench").Save(ctx)
					require.NoError(b, e)
					userIDs[i], keyIDs[i] = u.ID, k.ID
				}
				repo := &Store{}
				sequence := new(atomic.Int64)
				write := func() {
					current := sequence.Add(1)
					keys := make([]string, 64)
					prepared := make(map[string]usageLogInsertPrepared, 64)
					for i := range keys {
						log := &usage.UsageLog{UserID: userIDs[i], BillingUserID: userIDs[i], APIKeyID: keyIDs[i], ProviderID: p.ID, RequestID: fmt.Sprintf("bench-%d-%d", current, i), Model: "test", CreatedAt: time.Now(), InputTokens: 100, OutputTokens: 100, TotalCost: 0.1, ActualCost: 0.1}
						keys[i] = usageLogBatchKey(log.RequestID, log.APIKeyID)
						prepared[keys[i]] = prepareUsageLogInsert(log)
					}
					_, _, _, writeErr := repo.batchInsertUsageLogs(integrationDB, keys, prepared)
					if writeErr != nil {
						b.Error(writeErr)
					}
				}
				if concurrent {
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							write()
						}
					})
				} else {
					for b.Loop() {
						write()
					}
				}
			})
		}
	}
}
