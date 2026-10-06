package app

import (
	"log/slog"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/gateway/ws"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

// provideResponsesWSRuntime 装配共享参数快照，并在连接资源关闭前停止刷新。
func provideResponsesWSRuntime(cfg *config.Config, store *settings.Store, manager *lifecycle.Manager) *ws.Runtime {
	defaults := ws.DefaultParameters()
	if cfg != nil {
		v := cfg.Gateway.OpenAIWS
		defaults.MaxConnsPerProvider = v.MaxConnsPerProvider
		defaults.MinIdlePerProvider = v.MinIdlePerProvider
		defaults.MaxIdlePerProvider = v.MaxIdlePerProvider
		defaults.DynamicMaxConnsByProviderConcurrencyEnabled = v.DynamicMaxConnsByProviderConcurrencyEnabled
		defaults.OAuthMaxConnsFactor = v.OAuthMaxConnsFactor
		defaults.APIKeyMaxConnsFactor = v.APIKeyMaxConnsFactor
		defaults.QueueLimitPerConn = v.QueueLimitPerConn
		defaults.PoolTargetUtilization = v.PoolTargetUtilization
		defaults.PrewarmCooldownMS = v.PrewarmCooldownMS
		defaults.ClientFirstMessageTimeoutSeconds = v.ClientFirstMessageTimeoutSeconds
		defaults.IngressInterTurnIdleTimeoutSeconds = v.IngressInterTurnIdleTimeoutSeconds
		defaults.MaxIngressConnectionsPerAPIKey = v.MaxIngressConnectionsPerAPIKey
		defaults.ClientReadLimitBytes = v.ClientReadLimitBytes
		defaults.HTTPBridgeThresholdBytes = v.HTTPBridgeThresholdBytes
		defaults.DialTimeoutSeconds = v.DialTimeoutSeconds
		defaults.ReadTimeoutSeconds = v.ReadTimeoutSeconds
		defaults.WriteTimeoutSeconds = v.WriteTimeoutSeconds
		defaults.IngressPreviousResponseRecoveryEnabled = v.IngressPreviousResponseRecoveryEnabled
		defaults.StickySessionTTLSeconds = v.StickySessionTTLSeconds
		defaults.StickyResponseIDTTLSeconds = v.StickyResponseIDTTLSeconds
	}
	runtime := ws.NewRuntime(defaults, store, func(err error) { slog.Warn("Responses WS 设置刷新失败", "error", err) })
	manager.Register(lifecycle.Hook{Name: "ResponsesWSSettings", StartOrder: 980, StopOrder: 9, Start: runtime.Start, Stop: runtime.Stop})
	return runtime
}
