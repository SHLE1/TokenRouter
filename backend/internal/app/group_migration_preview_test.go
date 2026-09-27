package app

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/stretchr/testify/require"
)

// 预检必须列出未调用的自定义模型，并保持原提供商凭据配置不变。
func TestConfiguredModelsOutsideDefaults(t *testing.T) {
	value := &provider.Record{Platform: provider.PlatformOpenAI, Type: provider.ProviderTypeAPIKey, Credentials: map[string]any{
		"model_mapping":   map[string]any{"my-alias": "custom-upstream"},
		"model_whitelist": []string{"custom-upstream", "gpt-5.4"},
	}}
	require.ElementsMatch(t, []string{"custom-upstream", "my-alias"}, configuredModelsOutsideDefaults(value))
	require.Equal(t, []string{"custom-upstream", "gpt-5.4"}, value.Credentials["model_whitelist"])
	require.Contains(t, value.Credentials, "model_mapping")
	require.Empty(t, configuredModelsOutsideDefaults(&provider.Record{Platform: provider.PlatformOpenAI, Type: provider.ProviderTypeAPIKey}))
}
