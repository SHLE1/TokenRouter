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
