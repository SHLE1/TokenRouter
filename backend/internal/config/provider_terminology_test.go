package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// 显式为零或关闭的旧配置也要拒绝，不能被默认值掩盖。
func TestProviderTerminologyRejectsOldConfiguration(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigType("yaml")
	require.NoError(t, viper.ReadConfig(strings.NewReader("gateway:\n  max_account_switches: 0\n")))
	require.ErrorContains(t, rejectLegacyProviderConfig(), "max_provider_switches")
	viper.Reset()
	t.Setenv("GATEWAY_MAX_ACCOUNT_SWITCHES", "0")
	require.ErrorContains(t, rejectLegacyProviderConfig(), "GATEWAY_MAX_PROVIDER_SWITCHES")
}

// 配置名由用户定义，其中的 account 不应被当作旧领域字段。
func TestProviderTerminologyPreservesCustomProfileNames(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigType("yaml")
	require.NoError(t, viper.ReadConfig(strings.NewReader("gateway:\n  tls_fingerprint:\n    profiles:\n      account_gateway:\n        name: custom\n")))
	t.Setenv("GATEWAY_CUSTOM_ACCOUNT_LABEL", "custom")
	require.NoError(t, rejectLegacyProviderConfig())
}
