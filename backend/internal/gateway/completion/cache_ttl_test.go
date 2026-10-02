package completion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCacheOverrideProjection 检查缓存创建明细写回后，图片输入等其他计量值保持原值。
func TestCacheOverrideProjection(t *testing.T) {
	value := TokenUsage{InputTokens: 2, ImageInputTokens: 8, CacheCreationInputTokens: 12}
	require.False(t, applyCacheOverride(&value, "5m"))
	require.Equal(t, 12, value.CacheCreation5mTokens)
	require.True(t, applyCacheOverride(&value, "1h"))
	require.Zero(t, value.CacheCreation5mTokens)
	require.Equal(t, 12, value.CacheCreation1hTokens)
	require.Equal(t, 12, value.CacheCreationInputTokens)
	require.Equal(t, 8, value.ImageInputTokens)
	require.Equal(t, 2, value.InputTokens)
}
