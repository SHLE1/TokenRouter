package provider

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestMixedGroupProviderModelScope(t *testing.T) {
	// 空配置允许代理提供商接收任意型号，显式白名单限定最终模型。
	claude := "claude-sonnet-4-6"
	gpt := "gpt-5.4"
	for _, tc := range []struct {
		platform string
		own      string
		other    string
	}{
		{provider.PlatformAnthropic, claude, gpt},
		{provider.PlatformOpenAI, gpt, claude},
	} {
		t.Run(tc.platform, func(t *testing.T) {
			value := &provider.Record{Platform: tc.platform, Type: provider.ProviderTypeAPIKey}
			require.True(t, value.IsModelSupported(tc.own, ModelDefaults(), ModelRules(value)))
			require.True(t, value.IsModelSupported(tc.other, ModelDefaults(), ModelRules(value)))
			require.True(t, value.IsModelSupported("unknown-model", ModelDefaults(), ModelRules(value)))
			value.Credentials = map[string]any{"model_mapping": map[string]any{"custom-alias": "custom-upstream"}}
			require.True(t, value.IsModelSupported("custom-alias", ModelDefaults(), ModelRules(value)))
			value.Credentials["model_whitelist"] = []string{"custom-*"}
			require.True(t, value.IsModelSupported("custom-model", ModelDefaults(), ModelRules(value)))
			require.False(t, value.IsModelSupported(tc.own, ModelDefaults(), ModelRules(value)))
			require.NotContains(t, value.GetConfiguredRequestModels(ModelDefaults()), "custom-*")
			value.Credentials["model_whitelist"] = []string{"*"}
			require.True(t, value.IsModelSupported(tc.other, ModelDefaults(), ModelRules(value)))
			require.NotContains(t, value.GetConfiguredRequestModels(ModelDefaults()), "*")
		})
	}
}
