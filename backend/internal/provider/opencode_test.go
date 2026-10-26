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
	require.Equal(t, "https://opencode.ai/inference/go/google/v1beta", (ProtocolTarget{Record: r, Protocol: capability.ProtocolGeminiGenerateContent}).BaseURL())
	r.Credentials["api_base_urls"] = map[string]any{"gemini": "https://relay.example/google/v1beta"}
	require.Equal(t, "https://relay.example/google/v1beta", (ProtocolTarget{Record: r, Protocol: capability.ProtocolGeminiGenerateContent}).BaseURL())
	for _, field := range []string{"provider_mode", OpenCodeModelProtocolsKey, OpenCodeCatalogProtocolsKey, "api_base_urls", "base_url"} {
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

// TestOpenCodeAcceptedCatalogRoutesAndManualOverrides exercises a newly accepted
// catalog model and a manual override. Removing the manual entry restores the
// accepted suggestion, and secret rotation retains both separate configuration tables.
func TestOpenCodeAcceptedCatalogRoutesAndManualOverrides(t *testing.T) {
	r := &Record{Platform: PlatformOpenCode, Type: ProviderTypeAPIKey, Credentials: map[string]any{
		"api_key": "test", "provider_mode": "go",
		OpenCodeCatalogProtocolsKey: map[string]any{"go": map[string]any{"new-model": "anthropic_messages"}},
		OpenCodeModelProtocolsKey:   map[string]any{"go": map[string]any{"new-model": "systemone"}},
	}}
	require.NoError(t, ValidateOpenCodeCredentials(r))
	require.Equal(t, capability.ProtocolSystemOne, r.OpenCodeModelProtocols()["new-model"])
	r.Credentials[OpenCodeModelProtocolsKey] = map[string]any{}
	require.Equal(t, capability.ProtocolAnthropicMessages, r.OpenCodeModelProtocols()["new-model"])
	resolved, ok := r.RoutingSnapshot().ResolveModelRoute("new-model", "", nil)
	require.True(t, ok)
	require.Equal(t, capability.ProtocolAnthropicMessages, resolved.Protocol)
	rotated := PreserveProtocolCredentials(r.Credentials, map[string]any{"api_key": "rotated"})
	require.Equal(t, r.Credentials[OpenCodeCatalogProtocolsKey], rotated[OpenCodeCatalogProtocolsKey])
	require.Equal(t, r.Credentials[OpenCodeModelProtocolsKey], rotated[OpenCodeModelProtocolsKey])
	bad := CloneRecord(r)
	bad.Credentials[OpenCodeCatalogProtocolsKey] = map[string]any{"go": map[string]any{"new-model": "invalid"}}
	require.Error(t, ValidateOpenCodeCredentials(bad))
}

// TestOpenCodeResolvedRouteCopiesModeMappingAndMetadata changes configuration
// after taking a snapshot. The retained attempt binds one mapping hop to its
// payg encoding and author identity, while a fresh snapshot observes Go mode.
func TestOpenCodeResolvedRouteCopiesModeMappingAndMetadata(t *testing.T) {
	r := &Record{Platform: PlatformOpenCode, Type: ProviderTypeAPIKey, Credentials: map[string]any{
		"api_key": "test", "model_mapping": map[string]any{"alias-*": "minimax-m3", "minimax-m3": "wrong-second-hop"},
	}}
	snapshot := r.RoutingSnapshot()
	r.Credentials["provider_mode"] = "go"
	resolved, ok := snapshot.ResolveModelRoute("alias-fast", "", nil)
	require.True(t, ok)
	require.Equal(t, "minimax-m3", resolved.Model)
	require.Equal(t, "minimax/MiniMax-M3", resolved.AttributeModel)
	require.Equal(t, capability.ProtocolOpenAIChatCompletions, resolved.Protocol)
	fresh, ok := r.RoutingSnapshot().ResolveModelRoute("alias-fast", "", nil)
	require.True(t, ok)
	require.Equal(t, capability.ProtocolAnthropicMessages, fresh.Protocol)
	r.Credentials[UpstreamProtocolsKey] = []string{}
	_, ok = r.RoutingSnapshot().ResolveModelRoute("alias-fast", "", nil)
	require.False(t, ok)
}
