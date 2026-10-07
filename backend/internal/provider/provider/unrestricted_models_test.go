package provider

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/stretchr/testify/require"
)

// TestUnrestrictedModels 接受目录未知型号，白名单在映射之后限制最终型号。
func TestUnrestrictedModels(t *testing.T) {
	for _, platform := range []string{"anthropic", "openai", "gemini", "antigravity", "qoder", "grok", "kimi", "zhipu", "deepseek"} {
		t.Run(platform, func(t *testing.T) {
			value := &provider.Record{Platform: platform, Type: "apikey", Credentials: map[string]any{}}
			for _, explicit := range []bool{false, true} {
				if explicit {
					value.Credentials["model_whitelist"] = []any{}
				}
				require.True(t, value.IsModelSupported("unknown-new-model", ModelDefaults(), ModelRules(value)))
				require.True(t, value.HasUnrestrictedModelScope(ModelDefaults()))
			}
			value.Credentials["model_mapping"] = map[string]any{"alias": "unknown-new-model"}
			value.Credentials["model_whitelist"] = []any{"unknown-*"}
			require.True(t, value.IsModelSupported("alias", ModelDefaults(), ModelRules(value)))
			require.False(t, value.IsModelSupported("another-model", ModelDefaults(), ModelRules(value)))
			require.False(t, value.HasUnrestrictedModelScope(ModelDefaults()))
		})
	}
}

// TestSparkModelScope 保留 Spark 独立配额对应的型号限制。
func TestSparkModelScope(t *testing.T) {
	parent := int64(1)
	value := &provider.Record{Platform: "openai", Type: "oauth", ParentProviderID: &parent}
	require.False(t, value.HasUnrestrictedModelScope(ModelDefaults()))
	require.False(t, value.IsModelSupported("unknown-new-model", ModelDefaults(), ModelRules(value)))
	require.True(t, value.IsModelSupported("gpt-5.3-codex-spark", ModelDefaults(), ModelRules(value)))
}
