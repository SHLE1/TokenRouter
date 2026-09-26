//go:build integration

package bootstrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// 管理员并发升级独立于已下线的默认分组初始化，保留人工配置并保持幂等。
func TestSimpleModeAdminConcurrencyPreservesOverridesWithoutCreatingGroups(t *testing.T) {
	ctx := context.Background()
	client := testEntTx(t).Client()
	settings, err := client.Setting.Query().All(ctx)
	require.NoError(t, err)
	for _, item := range settings {
		if item.Key == "simple_mode_admin_concurrency_upgraded_30" {
			require.NoError(t, client.Setting.DeleteOneID(item.ID).Exec(ctx))
		}
	}
	legacy, err := client.User.Create().SetEmail("simple-legacy-admin@example.test").SetPasswordHash("hash").SetRole("admin").SetConcurrency(5).Save(ctx)
	require.NoError(t, err)
	configured, err := client.User.Create().SetEmail("simple-configured-admin@example.test").SetPasswordHash("hash").SetRole("admin").SetConcurrency(12).Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().SetEmail("simple-member@example.test").SetPasswordHash("hash").SetRole("user").SetConcurrency(5).Save(ctx)
	require.NoError(t, err)
	before, err := client.Group.Query().Count(ctx)
	require.NoError(t, err)
	require.NoError(t, ensureSimpleModeAdminConcurrency(ctx, client))
	legacy, err = client.User.Get(ctx, legacy.ID)
	require.NoError(t, err)
	require.Equal(t, 30, legacy.Concurrency)
	configured, err = client.User.Get(ctx, configured.ID)
	require.NoError(t, err)
	require.Equal(t, 12, configured.Concurrency)
	member, err = client.User.Get(ctx, member.ID)
	require.NoError(t, err)
	require.Equal(t, 5, member.Concurrency)
	_, err = client.User.UpdateOneID(legacy.ID).SetConcurrency(7).Save(ctx)
	require.NoError(t, err)
	require.NoError(t, ensureSimpleModeAdminConcurrency(ctx, client))
	legacy, err = client.User.Get(ctx, legacy.ID)
	require.NoError(t, err)
	require.Equal(t, 7, legacy.Concurrency)
	after, err := client.Group.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
}
