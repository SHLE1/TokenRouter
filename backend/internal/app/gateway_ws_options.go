package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	wshttp "github.com/TokenFlux/TokenRouter/internal/gateway/ws/httpapi"
)

// openAIWSExecutionOptions 读取 WS 执行参数，处理 nil 配置并传递配置中的零值。
func openAIWSExecutionOptions(cfg *config.Config) *wshttp.OpenAIWSOptions {
	if cfg == nil {
		return nil
	}
	v := cfg.Gateway.OpenAIWS
	return &wshttp.OpenAIWSOptions{
		MaxConnsPerProvider:                         v.MaxConnsPerProvider,
		MinIdlePerProvider:                          v.MinIdlePerProvider,
		MaxIdlePerProvider:                          v.MaxIdlePerProvider,
		DynamicMaxConnsByProviderConcurrencyEnabled: v.DynamicMaxConnsByProviderConcurrencyEnabled,
		OAuthMaxConnsFactor:                         v.OAuthMaxConnsFactor,
		APIKeyMaxConnsFactor:                        v.APIKeyMaxConnsFactor,
		QueueLimitPerConn:                           v.QueueLimitPerConn,
		PoolTargetUtilization:                       v.PoolTargetUtilization,
		PrewarmCooldownMS:                           v.PrewarmCooldownMS,
		ClientFirstMessageTimeoutSeconds:            v.ClientFirstMessageTimeoutSeconds,
		IngressInterTurnIdleTimeoutSeconds:          v.IngressInterTurnIdleTimeoutSeconds,
		MaxIngressConnectionsPerAPIKey:              v.MaxIngressConnectionsPerAPIKey,
		ClientReadLimitBytes:                        v.ClientReadLimitBytes,
		HTTPBridgeThresholdBytes:                    v.HTTPBridgeThresholdBytes,
		DialTimeoutSeconds:                          v.DialTimeoutSeconds,
		ReadTimeoutSeconds:                          v.ReadTimeoutSeconds,
		WriteTimeoutSeconds:                         v.WriteTimeoutSeconds,
		IngressPreviousResponseRecoveryEnabled:      v.IngressPreviousResponseRecoveryEnabled,
		StickySessionTTLSeconds:                     v.StickySessionTTLSeconds,
		StickyResponseIDTTLSeconds:                  v.StickyResponseIDTTLSeconds,
	}
}
