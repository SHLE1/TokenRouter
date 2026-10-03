package usage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUsageStatsCacheKey_StableAndDistinct(t *testing.T) {
	start := time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	base := UsageLogFilters{StartTime: &start, EndTime: &end, Model: "claude-3"}

	k1 := usageStatsCacheKey(base)
	k2 := usageStatsCacheKey(base)
	require.NotEmpty(t, k1)
	require.Equal(t, k1, k2, "same filters must produce same key")

	other := base
	other.Model = "gpt-4o"
	require.NotEqual(t, k1, usageStatsCacheKey(other), "different model must change key")

	withUser := base
	withUser.UserID = 7
	require.NotEqual(t, k1, usageStatsCacheKey(withUser), "different user must change key")

	// 团队维度必须参与缓存键，防止切换团队后复用旧汇总。
	withTeam := base
	withTeam.TeamID = 11
	require.NotEqual(t, k1, usageStatsCacheKey(withTeam), "different team must change key")
	// 同一范围的不同端点图分别缓存。
	keys := map[string]bool{k1: true}
	for _, source := range []string{"inbound", "upstream", "path"} {
		filtered := base
		filtered.EndpointSource = source
		key := usageStatsCacheKey(filtered)
		require.False(t, keys[key])
		keys[key] = true
	}
}
