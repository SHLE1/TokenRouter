//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
	"github.com/TokenFlux/TokenRouter/internal/testutil/postgrescontainer"
)

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
