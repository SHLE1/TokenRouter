package provider

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// TestOpenCodeCredentialsValidateAndResolveMode exercises save validation and
// endpoint selection for API-key providers. Per-mode model overrides and custom
// protocol bases remain isolated from another mode's defaults.
func TestOpenCodeCredentialsValidateAndResolveMode(t *testing.T) {
	r := &Record{Platform: PlatformOpenCode, Type: ProviderTypeAPIKey, Credentials: map[string]any{"api_key": "test"}}
	require.NoError(t, NormalizeProviderProtocols(r))
	require.Len(t, r.UpstreamProtocols(), 5)
	require.Equal(t, "https://opencode.ai/inference/openai/v1", r.OpenCodeBaseURL(capability.ProtocolOpenAIChatCompletions))
	r.Credentials["provider_mode"] = ProviderModeGo
	r.Credentials[OpenCodeModelProtocolsKey] = map[string]any{"go": map[string]any{"custom": "systemone"}, "payg": map[string]any{"custom": "anthropic_messages"}}
	require.NoError(t, ValidateOpenCodeCredentials(r))
	require.Equal(t, capability.ProtocolSystemOne, r.OpenCodeModelProtocols()["custom"])
	require.Equal(t, "https://opencode.ai/inference/go/google", r.GetGeminiBaseURL("unused"))
	r.Credentials["api_base_urls"] = map[string]any{"gemini": "https://relay.example/google/v1beta"}
	require.Equal(t, "https://relay.example/google", r.GetGeminiBaseURL("unused"))
	for _, field := range []string{"provider_mode", OpenCodeModelProtocolsKey, "api_base_urls", "base_url"} {
		invalid := CloneRecord(r)
		invalid.Credentials[field] = 7
		require.Error(t, ValidateOpenCodeCredentials(invalid), field)
	}
	r.Type = ProviderTypeOAuth
	require.Error(t, ValidateOpenCodeCredentials(r))
}

// TestOpenCodeUsageModesKeepPaygDisabled supplies an attempted manual query
// override on payg. Only Go can select its fixed adapter, and configured generic
// adapter text cannot change the OpenCode response interpretation.
func TestOpenCodeUsageModesKeepPaygDisabled(t *testing.T) {
	r := &Record{Platform: PlatformOpenCode, Type: ProviderTypeAPIKey, Credentials: map[string]any{}, Extra: map[string]any{UpstreamUsageQueryExtraKey: map[string]any{"enabled": true, "adapter": "new_api"}}}
	config, err := EffectiveUpstreamUsageConfig(r)
	require.NoError(t, err)
	require.False(t, config.Enabled)
	r.Credentials["provider_mode"] = ProviderModeGo
	config, err = EffectiveUpstreamUsageConfig(r)
	require.NoError(t, err)
	require.True(t, config.Enabled)
	require.Equal(t, UpstreamUsageAdapterOpenCodeGo, config.Adapter)
}
