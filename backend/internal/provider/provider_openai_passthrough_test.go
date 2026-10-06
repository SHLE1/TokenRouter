package provider_test

import (
	"testing"

	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

func TestProvider_IsOpenAIPassthroughEnabled(t *testing.T) {
	t.Run("新字段开启", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		}
		require.True(t, provider.IsOpenAIPassthroughEnabled())
	})

	t.Run("兼容旧字段", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"openai_oauth_passthrough": true,
			},
		}
		require.True(t, provider.IsOpenAIPassthroughEnabled())
	})

	t.Run("非OpenAI提供商始终关闭", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformAnthropic,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		}
		require.False(t, provider.IsOpenAIPassthroughEnabled())
	})

	t.Run("空额外配置默认关闭", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
		}
		require.False(t, provider.IsOpenAIPassthroughEnabled())
	})
}

func TestProvider_IsOpenAIOAuthPassthroughEnabled(t *testing.T) {
	t.Run("仅OAuth类型允许返回开启", func(t *testing.T) {
		oauthProvider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		}
		require.True(t, oauthProvider.IsOpenAIOAuthPassthroughEnabled())

		apiKeyProvider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Extra: map[string]any{
				"openai_passthrough": true,
			},
		}
		require.False(t, apiKeyProvider.IsOpenAIOAuthPassthroughEnabled())
	})
}

func TestProvider_IsCodexCLIOnlyEnabled(t *testing.T) {
	t.Run("OpenAI OAuth 开启", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"codex_cli_only": true,
			},
		}
		require.True(t, provider.IsCodexCLIOnlyEnabled())
	})

	t.Run("OpenAI OAuth 关闭", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"codex_cli_only": false,
			},
		}
		require.False(t, provider.IsCodexCLIOnlyEnabled())
	})

	t.Run("字段缺失默认关闭", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra:    map[string]any{},
		}
		require.False(t, provider.IsCodexCLIOnlyEnabled())
	})

	t.Run("类型非法默认关闭", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"codex_cli_only": "true",
			},
		}
		require.False(t, provider.IsCodexCLIOnlyEnabled())
	})

	t.Run("非 OAuth 提供商始终关闭", func(t *testing.T) {
		apiKeyProvider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Extra: map[string]any{
				"codex_cli_only": true,
			},
		}
		require.False(t, apiKeyProvider.IsCodexCLIOnlyEnabled())

		otherPlatform := &providercore.Record{
			Platform: capability.PlatformAnthropic,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"codex_cli_only": true,
			},
		}
		require.False(t, otherPlatform.IsCodexCLIOnlyEnabled())
	})

	t.Run("新策略字段优先于旧字段", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"openai_oauth_client_policy": providercore.OpenAIOAuthClientPolicyAny,
				"codex_cli_only":             true,
			},
		}
		require.False(t, provider.IsCodexCLIOnlyEnabled())
		require.Equal(t, providercore.OpenAIOAuthClientPolicyAny, provider.GetOpenAIOAuthClientPolicy())
	})

	t.Run("TLS 路由器策略不等同于 Codex-only", func(t *testing.T) {
		provider := &providercore.Record{
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeOAuth,
			Extra: map[string]any{
				"openai_oauth_client_policy": providercore.OpenAIOAuthClientPolicyTLSRouterMatchedOnly,
				"tls_fingerprint_router_id":  int64(12),
			},
		}
		require.False(t, provider.IsCodexCLIOnlyEnabled())
		require.True(t, provider.IsOpenAIOAuthTLSRouterMatchedOnly())
		require.Equal(t, int64(12), provider.GetTLSFingerprintRouterID())
	})
}

func TestProvider_IsTLSFingerprintEnabled(t *testing.T) {
	tests := []struct {
		name     string
		provider *providercore.Record
		want     bool
	}{
		{
			name: "Anthropic OAuth 开启",
			provider: &providercore.Record{
				Platform: capability.PlatformAnthropic,
				Type:     capability.ProviderTypeOAuth,
				Extra:    map[string]any{"enable_tls_fingerprint": true},
			},
			want: true,
		},
		{
			name: "Anthropic SetupToken 开启",
			provider: &providercore.Record{
				Platform: capability.PlatformAnthropic,
				Type:     capability.ProviderTypeSetupToken,
				Extra:    map[string]any{"enable_tls_fingerprint": true},
			},
			want: true,
		},
		{
			name: "OpenAI OAuth 开启",
			provider: &providercore.Record{
				Platform: capability.PlatformOpenAI,
				Type:     capability.ProviderTypeOAuth,
				Extra:    map[string]any{"enable_tls_fingerprint": true},
			},
			want: true,
		},
		{
			name: "OpenAI API Key 不支持",
			provider: &providercore.Record{
				Platform: capability.PlatformOpenAI,
				Type:     capability.ProviderTypeAPIKey,
				Extra:    map[string]any{"enable_tls_fingerprint": true},
			},
			want: false,
		},
		{
			name: "非法类型按关闭处理",
			provider: &providercore.Record{
				Platform: capability.PlatformOpenAI,
				Type:     capability.ProviderTypeOAuth,
				Extra:    map[string]any{"enable_tls_fingerprint": "true"},
			},
			want: false,
		},
		{
			name: "字段缺失按关闭处理",
			provider: &providercore.Record{
				Platform: capability.PlatformOpenAI,
				Type:     capability.ProviderTypeOAuth,
				Extra:    map[string]any{},
			},
			want: false,
		},
		{
			name:     "nil 提供商按关闭处理",
			provider: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.provider.IsTLSFingerprintEnabled())
		})
	}
}
