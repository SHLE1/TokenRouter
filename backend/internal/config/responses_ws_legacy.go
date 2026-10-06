package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// warnLegacyResponsesWSSettings 提示部署者从分组协议控制客户端 WS 权限。
func warnLegacyResponsesWSSettings() {
	for _, key := range []string{"gateway.openai_ws.allow_store_recovery", "gateway.openai_ws.apikey_enabled", "gateway.openai_ws.enabled", "gateway.openai_ws.event_flush_batch_size", "gateway.openai_ws.event_flush_interval_ms", "gateway.openai_ws.force_http", "gateway.openai_ws.http_bridge_enabled", "gateway.openai_ws.ingress_mode_default", "gateway.openai_ws.mode_router_v2_enabled", "gateway.openai_ws.oauth_enabled", "gateway.openai_ws.payload_log_sample_rate", "gateway.openai_ws.prewarm_generate_enabled", "gateway.openai_ws.responses_websockets", "gateway.openai_ws.responses_websockets_v2", "gateway.openai_ws.retry_backoff_initial_ms", "gateway.openai_ws.retry_backoff_max_ms", "gateway.openai_ws.retry_jitter_ratio", "gateway.openai_ws.retry_total_budget_ms", "gateway.openai_ws.store_disabled_conn_mode", "gateway.openai_ws.store_disabled_force_new_conn"} {
		_ = viper.BindEnv(key)
		if viper.IsSet(key) {
			slog.Warn("旧 Responses WS 参数已停用，请在分组协议中设置客户端权限，在网关设置中调整连接参数", "key", key)
		}
	}
}
