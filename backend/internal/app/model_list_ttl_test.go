package app

import (
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/stretchr/testify/require"
)

// TestGatewayHotpathHelpers_CacheTTLAndStickyContext 检查 app 装配的默认缓存 TTL、自定义 TTL 和粘性会话参数。
func TestGatewayHotpathHelpers_CacheTTLAndStickyContext(t *testing.T) {
	t.Run("resolve_models_list_cache_ttl", func(t *testing.T) {
		require.Equal(t, 15*time.Second, resolveModelsListCacheTTL(nil))

		cfg := &config.Config{
			Gateway: config.GatewayConfig{
				ModelsListCacheTTLSeconds: 20,
			},
		}
		require.Equal(t, 20*time.Second, resolveModelsListCacheTTL(cfg))
	})
}
