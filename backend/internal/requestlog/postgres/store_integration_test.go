//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
	"github.com/TokenFlux/TokenRouter/internal/requestlog"
	"github.com/TokenFlux/TokenRouter/internal/testutil/postgrescontainer"
)

// lookupPlanNode 读取 PostgreSQL 执行计划中的扫描方式与表名。
type lookupPlanNode struct {
	Type     string           `json:"Node Type"`
	Relation string           `json:"Relation Name"`
	Plans    []lookupPlanNode `json:"Plans"`
}

// TestStoreLookupIsolationAndReplay 在实际迁移后的 PostgreSQL 上检查别名、归属和重放。
func TestStoreLookupIsolationAndReplay(t *testing.T) {
	db := postgrescontainer.New(t)
	store := NewStore(db)
	ctx := context.Background()
	start := time.Now().UTC().Truncate(time.Microsecond)
	first := telemetry.RequestRecord{RequestID: "first", UserID: 101, State: "running", StartedAt: start, UpdatedAt: start, Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}}}
	second := telemetry.RequestRecord{RequestID: "second", UserID: 202, State: "failed", StartedAt: start, UpdatedAt: start, Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}}}
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{first, second}))
	items, err := store.Find(ctx, "shared", 101, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "first", items[0].RequestID)
	items, err = store.Find(ctx, "shared", 0, true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	items, err = store.Find(ctx, "first", 202, false)
	require.NoError(t, err)
	require.Empty(t, items)
	completed := first
	completed.UpdatedAt = start.Add(time.Second)
	completed.State = "completed"
	completed.Aliases = []telemetry.RequestAlias{{Kind: "billing", Value: "client:first"}}
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{completed}))
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{first}))
	items, err = store.Find(ctx, "client:first", 101, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "completed", items[0].State)
	items, err = store.Find(ctx, "shared", 101, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	// 纳秒版本号独立于 PostgreSQL 时间列的微秒精度。
	completed.UpdatedAt = completed.UpdatedAt.Add(time.Nanosecond)
	completed.Model = "latest"
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{completed}))
	items, err = store.Find(ctx, "first", 101, false)
	require.NoError(t, err)
	require.Equal(t, "latest", items[0].Model)

	_, err = db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash) VALUES
      (101,'member@requests.test','hash'),(202,'next-owner@requests.test','hash'),(303,'owner@requests.test','hash')`)
	require.NoError(t, err)
	var teamID int64
	require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO teams(name) VALUES('requests') RETURNING id`).Scan(&teamID))
	_, err = db.ExecContext(ctx, `INSERT INTO team_memberships(team_id,user_id,role) VALUES($1,101,'member'),($1,303,'owner')`, teamID)
	require.NoError(t, err)
	teamRecord := telemetry.RequestRecord{RequestID: "team-request", UserID: 101, TeamID: teamID, State: "failed", StartedAt: start, UpdatedAt: start}
	unknown := telemetry.RequestRecord{RequestID: "unknown", TeamID: teamID, State: "failed", StartedAt: start, UpdatedAt: start}
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{teamRecord, unknown}))
	items, err = store.Find(ctx, "team-request", 303, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	items, err = store.Find(ctx, "unknown", 303, false)
	require.NoError(t, err)
	require.Empty(t, items)
	items, err = store.Find(ctx, "first", 303, false)
	require.NoError(t, err)
	require.Empty(t, items)
	_, err = db.ExecContext(ctx, `UPDATE team_memberships SET role='member' WHERE user_id=303;
      INSERT INTO team_memberships(team_id,user_id,role) SELECT team_id,202,'owner' FROM team_memberships WHERE user_id=303`)
	require.NoError(t, err)
	items, err = store.Find(ctx, "team-request", 303, false)
	require.NoError(t, err)
	require.Empty(t, items)
	items, err = store.Find(ctx, "team-request", 202, false)
	require.NoError(t, err)
	require.Len(t, items, 1)

	require.NoError(t, store.Cleanup(ctx, start.Add(24*time.Hour)))
	items, err = store.Find(ctx, "first", 0, true)
	require.NoError(t, err)
	require.Empty(t, items)
}

// TestStoreBatchWriterFlushesFinalSnapshot 在 PostgreSQL 上确认关闭排空和同一请求的版本合并。
func TestStoreBatchWriterFlushesFinalSnapshot(t *testing.T) {
	store := NewStore(postgrescontainer.New(t))
	service := requestlog.NewService(store, 30, nil)
	start := time.Now().UTC().Truncate(time.Microsecond)
	service.Observe(telemetry.RequestRecord{
		RequestID: "batched", UserID: 101, State: "running", StartedAt: start, UpdatedAt: start,
		Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "client-alias"}},
	})
	finished := start.Add(time.Second)
	service.Observe(telemetry.RequestRecord{
		RequestID: "batched", UserID: 101, State: "failed", UpdatedAt: finished, FinishedAt: &finished,
		Status: 502, Attempts: []telemetry.RequestAttempt{{Number: 1, ProviderID: 7, Status: 502}},
	})
	require.NoError(t, service.Stop(t.Context()))
	items, err := store.Find(t.Context(), "client-alias", 101, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "failed", items[0].State)
	require.Equal(t, 502, items[0].Status)
	require.Equal(t, start, items[0].StartedAt)
	require.Len(t, items[0].Attempts, 1)
	health, err := service.Health()
	require.NoError(t, err)
	require.Zero(t, health.Pending)
	require.Zero(t, health.Failures)
}

// TestStoreLookupIndexesAndAliases 检查历史标识、用户归属和稀疏 ID 查询的索引扫描。
func TestStoreLookupIndexesAndAliases(t *testing.T) {
	db := postgrescontainer.New(t)
	ctx := t.Context()
	_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,password_hash) VALUES (101,'lookup@requests.test','hash'),(202,'other@requests.test','hash');
INSERT INTO api_keys(id,user_id,key,name) VALUES (101,101,'lookup-key','lookup'),(202,202,'other-key','other');
INSERT INTO providers(id,name,platform,type) VALUES (101,'lookup','openai','apikey');
INSERT INTO usage_logs(user_id,billing_user_id,api_key_id,provider_id,request_id,billing_key,upstream_request_id,model)
VALUES (101,101,101,101,'new-request','client:billing','supplier','model'),
       (202,202,202,101,'other-request','client:billing','supplier','model'),
       (101,101,101,101,'client:legacy',NULL,NULL,'model');
INSERT INTO ops_error_logs(request_id,client_request_id,user_id,error_phase,error_type,status_code)
SELECT 'error-'||g, 'error-client-'||g, 101,'upstream','api_error',502 FROM generate_series(1,20000) g;
INSERT INTO ops_system_logs(request_id,client_request_id,level,message)
SELECT 'system-'||g, 'system-client-'||g,'info','lookup fixture' FROM generate_series(1,20000) g;
INSERT INTO ops_system_logs(request_id,client_request_id,level,message) VALUES ('client:legacy','legacy','info','legacy fixture');
INSERT INTO audit_logs(request_id,actor_user_id,status_code) VALUES ('new-request',101,200);
ANALYZE ops_error_logs;
ANALYZE ops_system_logs;
ANALYZE usage_logs;`)
	require.NoError(t, err)
	store := NewStore(db)
	now := time.Now().UTC()
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{
		{RequestID: "new-request", UserID: 101, APIKeyID: 101, StartedAt: now, UpdatedAt: now, State: "completed", Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}, {Kind: "billing", Value: "client:billing"}, {Kind: "legacy", Value: "client:legacy"}, {Kind: "related", Value: "child"}}},
		{RequestID: "other-request", UserID: 202, APIKeyID: 202, StartedAt: now, UpdatedAt: now, State: "completed", Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}}},
		{RequestID: "child", ParentRequestID: "new-request", UserID: 101, StartedAt: now, UpdatedAt: now, State: "failed"},
	}))
	for _, test := range []struct {
		search    string
		candidate string
		visible   []string
	}{
		{"new-request", "new-request", []string{"new-request", "child", "client:legacy"}},
		{"shared", "new-request", []string{"new-request", "child", "client:legacy"}},
		{"supplier", "new-request", []string{"new-request", "child", "client:legacy"}},
		{"client:billing", "new-request", []string{"new-request", "child", "client:legacy"}},
		{"legacy", "client:legacy", []string{"new-request", "child", "client:legacy"}},
		{"client:legacy", "client:legacy", []string{"new-request", "child", "client:legacy"}},
		{"generated:legacy", "client:legacy", []string{"new-request", "child", "client:legacy"}},
		{"error-1", "error-1", []string{"error-1"}},
		{"error-client-1", "error-1", []string{"error-1"}},
		{"system-1", "system-1", nil},
		{"system-client-1", "system-1", nil},
		{"missing", "missing", nil},
	} {
		t.Run(test.search, func(t *testing.T) {
			var candidates []string
			require.NoError(t, db.QueryRowContext(ctx, `SELECT ARRAY(SELECT id FROM request_lookup_ids($1))`, test.search).Scan(pq.Array(&candidates)))
			require.Contains(t, candidates, test.candidate)
			items, err := store.Find(ctx, test.search, 101, false)
			require.NoError(t, err)
			ids := make([]string, 0, len(items))
			for _, item := range items {
				ids = append(ids, item.RequestID)
			}
			require.ElementsMatch(t, test.visible, ids)
		})
	}
	items, err := store.Find(ctx, "shared", 101, false)
	require.NoError(t, err)
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.RequestID)
	}
	require.ElementsMatch(t, []string{"new-request", "child", "client:legacy"}, ids)
	items, err = store.Find(ctx, "shared", 0, true)
	require.NoError(t, err)
	require.Len(t, items, 4)

	// 执行计划断言避免把机器速度当作性能门禁，两个历史表都应使用 ID 索引。
	for _, id := range []string{"error-1", "system-client-1", "missing"} {
		var raw []byte
		require.NoError(t, db.QueryRowContext(ctx, `EXPLAIN (ANALYZE, FORMAT JSON) SELECT id FROM request_lookup_ids($1)`, id).Scan(&raw))
		var plans []struct {
			Plan lookupPlanNode `json:"Plan"`
		}
		require.NoError(t, json.Unmarshal(raw, &plans))
		seen := map[string]bool{}
		var inspect func(lookupPlanNode)
		inspect = func(node lookupPlanNode) {
			if node.Relation == "ops_system_logs" || node.Relation == "ops_error_logs" {
				require.NotEqual(t, "Seq Scan", node.Type, "%s: %s", id, string(raw))
				seen[node.Relation] = true
			}
			for _, child := range node.Plans {
				inspect(child)
			}
		}
		inspect(plans[0].Plan)
		require.True(t, seen["ops_system_logs"])
		require.True(t, seen["ops_error_logs"])
	}
}

// TestStoreExactIDIgnoresExternalCollisions 检查共享调用方 ID 和上游 ID 同名时的精确查询。
func TestStoreExactIDIgnoresExternalCollisions(t *testing.T) {
	db := postgrescontainer.New(t)
	ctx := t.Context()
	_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,password_hash) VALUES (101,'exact@requests.test','hash'),(202,'collision@requests.test','hash');
INSERT INTO api_keys(id,user_id,key,name) VALUES (101,101,'exact-key','exact'),(202,202,'collision-key','collision');
INSERT INTO providers(id,name,platform,type) VALUES (101,'exact','openai','apikey');
INSERT INTO usage_logs(user_id,billing_user_id,api_key_id,provider_id,request_id,upstream_request_id,model)
VALUES (101,101,101,101,'local-first','supplier-first','model'),(202,202,202,101,'local-second','local-first','model');
INSERT INTO ops_system_logs(request_id,client_request_id,level,message)
VALUES ('local-first','shared','info','first'),('local-second','shared','info','second');
INSERT INTO ops_error_logs(request_id,client_request_id,user_id,error_phase,error_type,status_code)
VALUES ('local-first','shared',101,'upstream','api_error',502),
       ('local-second','shared',202,'upstream','api_error',502),
       ('local-second','local-first',202,'upstream','api_error',502);
INSERT INTO audit_logs(request_id,actor_user_id,status_code)
VALUES ('local-first',101,502),('local-second',202,502);`)
	require.NoError(t, err)
	store := NewStore(db)
	now := time.Now().UTC()
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{
		{RequestID: "local-first", UserID: 101, APIKeyID: 101, StartedAt: now, UpdatedAt: now, State: "failed", Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}}},
		{RequestID: "local-second", UserID: 202, APIKeyID: 202, StartedAt: now, UpdatedAt: now, State: "failed", Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}, {Kind: "caller", Value: "local-first"}}},
	}))
	var candidates []string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT ARRAY(SELECT id FROM request_lookup_ids('local-first'))`).Scan(pq.Array(&candidates)))
	require.Equal(t, []string{"local-first"}, candidates)
	for _, search := range []string{"local-first", " local-first "} {
		items, err := store.Find(ctx, search, 0, true)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, "local-first", items[0].RequestID)
		require.Len(t, items[0].Usage, 1)
		require.Len(t, items[0].Errors, 1)
		require.Len(t, items[0].AuditIDs, 1)
	}
	items, err := store.Find(ctx, "local-first", 202, false)
	require.NoError(t, err)
	require.Empty(t, items)
	items, err = store.Find(ctx, "shared", 0, true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	items, err = store.Find(ctx, "shared", 101, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "local-first", items[0].RequestID)

	// 父子关联来自请求摘要，查询子请求时保持自己的 ID 范围。
	require.NoError(t, store.Save(ctx, []telemetry.RequestRecord{{
		RequestID: "child", ParentRequestID: "local-first", UserID: 101, StartedAt: now, UpdatedAt: now,
		State: "completed", Aliases: []telemetry.RequestAlias{{Kind: "caller", Value: "shared"}},
	}}))
	items, err = store.Find(ctx, "local-first", 0, true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	for _, item := range items {
		require.Contains(t, []string{"local-first", "child"}, item.RequestID)
	}
	items, err = store.Find(ctx, "child", 0, true)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "child", items[0].RequestID)
}
