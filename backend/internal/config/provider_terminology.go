package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// rejectLegacyProviderConfig 拒绝旧部署键，避免 Viper 忽略旧值后使用默认限额。
func rejectLegacyProviderConfig() error {
	// 只拒绝已更名的配置项，TLS 配置名等用户定义的键不属于领域字段。
	keys := []string{
		"gateway.max_account_switches",
		"gateway.max_account_switches_gemini",
		"gateway.openai_ws.max_conns_per_account",
		"gateway.openai_ws.min_idle_per_account",
		"gateway.openai_ws.max_idle_per_account",
		"gateway.openai_ws.dynamic_max_conns_by_account_concurrency_enabled",
	}
	for _, key := range keys {
		if viper.InConfig(key) {
			return fmt.Errorf("configuration %q has been renamed to %q", key, strings.ReplaceAll(key, "account", "provider"))
		}
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if _, exists := os.LookupEnv(envKey); exists {
			return fmt.Errorf("environment variable %q has been renamed to %q", envKey, strings.ReplaceAll(envKey, "ACCOUNT", "PROVIDER"))
		}
	}
	return nil
}
