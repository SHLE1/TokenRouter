package app

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/stretchr/testify/require"
)

// TestResponseHeaderFilterConfigurationProjection 检查响应头过滤配置的 nil、默认值、增删规则及输入副本隔离。
func TestResponseHeaderFilterConfigurationProjection(t *testing.T) {
	require.Nil(t, provideResponseHeaderFilter(nil))
	cfg := &config.Config{}
	cfg.Security.ResponseHeaders.AdditionalAllowed = []string{"X-Trace-Custom"}
	cfg.Security.ResponseHeaders.ForceRemove = []string{"X-Request-Id"}
	disabled := provideResponseHeaderFilter(cfg)
	require.True(t, disabled.Allows("X-Request-Id"))
	require.False(t, disabled.Allows("X-Trace-Custom"))
	cfg.Security.ResponseHeaders.Enabled = true
	enabled := provideResponseHeaderFilter(cfg)
	require.False(t, enabled.Allows("X-Request-Id"))
	require.True(t, enabled.Allows("X-Trace-Custom"))
	require.False(t, enabled.Allows("Connection"))
	cfg.Security.ResponseHeaders.AdditionalAllowed[0] = "X-Replaced"
	cfg.Security.ResponseHeaders.ForceRemove[0] = "Content-Type"
	require.True(t, enabled.Allows("X-Trace-Custom"))
	require.True(t, enabled.Allows("Content-Type"))
	require.False(t, enabled.Allows("X-Replaced"))
}
