package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// rejectLegacyProviderConfig 拒绝旧部署键，避免 Viper 忽略旧值后使用默认限额。
func rejectLegacyProviderConfig() error {
	for _, key := range viper.AllKeys() {
		if strings.Contains(key, "account") && viper.InConfig(key) {
			return fmt.Errorf("configuration %q has been renamed to %q", key, strings.ReplaceAll(key, "account", "provider"))
		}
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if (strings.HasPrefix(key, "GATEWAY_") || strings.HasPrefix(key, "TOKEN_REFRESH_")) && strings.Contains(key, "ACCOUNT") {
			return fmt.Errorf("environment variable %q has been renamed to %q", key, strings.ReplaceAll(key, "ACCOUNT", "PROVIDER"))
		}
	}
	return nil
}
