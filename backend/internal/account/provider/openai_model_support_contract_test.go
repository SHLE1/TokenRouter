//go:build unit

package provider

import (
	"testing"

	acct "github.com/TokenFlux/TokenRouter/internal/account"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

func newOpenAIOAuthAccountForModelTest() *acct.Record {
	return &acct.Record{
		ID:       1,
		Platform: capability.PlatformOpenAI,
		Type:     capability.AccountTypeOAuth,
	}
}

func TestIsModelSupported_OpenAIOAuthEmptyMapping_UsesDefaultDirectory(t *testing.T) {
	account := newOpenAIOAuthAccountForModelTest()
	for _, model := range []string{"gpt-5.4", "gpt-5.6-terra", "gpt-5.6-sol"} {
		require.True(t, account.IsModelSupported(model, ModelDefaults(), ModelRules(account)), model)
	}
	for _, model := range []string{"", "my-custom-alias", "model-outside-current-catalog"} {
		require.False(t, account.IsModelSupported(model, ModelDefaults(), ModelRules(account)), model)
	}
}

func TestIsModelSupported_OpenAIOAuthEmptyMapping_RejectsForeignModels(t *testing.T) {
	account := newOpenAIOAuthAccountForModelTest()

	// Codex 上游必然以不可重试的 400 拒绝这些厂商家族，调度阶段就应跳过
	// 该账号，让明确支持的 API Key 账号接手。
	foreign := []string{
		"deepseek-v4",
		"deepseek-chat",
		"glm-4.7",
		"kimi-k2",
		"k3",
		"k3-256k",
		"moonshot-v1-128k",
		"gemini-3.0-pro",
		"grok-4",
		"qwen3-max",
		"minimax-m2.5",
		"llama-3.3-70b",
		"provider/deepseek-v4", // vendor/model 形式取最后一段判定。
		"provider/k3",          // Kimi Code bare ID 的 vendor/model 形式。
	}
	for _, model := range foreign {
		require.False(t, account.IsModelSupported(model, ModelDefaults(), ModelRules(account)), "expected %q to be rejected by empty-mapping OpenAI OAuth account", model)
	}
}

func TestIsModelSupported_OpenAIOAuthMappingKeepsForkSemantics(t *testing.T) {
	account := newOpenAIOAuthAccountForModelTest()
	account.Credentials = map[string]any{
		"model_mapping": map[string]any{"deepseek-v4": "gpt-5.4", "k3": "gpt-5.4"},
	}

	// 映射可以引入别名；未命中的模型仍受默认目录和认证能力限制。
	require.True(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
	require.True(t, account.IsModelSupported("k3", ModelDefaults(), ModelRules(account)))
	require.False(t, account.IsModelSupported("glm-4.7", ModelDefaults(), ModelRules(account)))

	account.Credentials["model_whitelist"] = []any{"gpt-5.4"}
	require.True(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
	require.False(t, account.IsModelSupported("glm-4.7", ModelDefaults(), ModelRules(account)))
}

func TestIsModelSupported_OpenAIOAuthEmptyMappingRespectsWhitelist(t *testing.T) {
	account := newOpenAIOAuthAccountForModelTest()
	account.Credentials = map[string]any{"model_whitelist": []any{"gpt-5.4", "deepseek-v4"}}

	// 最终白名单仍先限制允许范围，但不能让 Codex 上游无法服务的厂商模型绕过平台限制。
	require.True(t, account.IsModelSupported("gpt-5.4", ModelDefaults(), ModelRules(account)))
	require.False(t, account.IsModelSupported("gpt-5.3-codex", ModelDefaults(), ModelRules(account)))
	require.False(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
}

func TestIsModelSupported_OpenAIOAuthPassthroughKeepsModelScope(t *testing.T) {
	account := newOpenAIOAuthAccountForModelTest()
	account.Extra = map[string]any{"openai_passthrough": true}
	require.False(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
	account.Credentials = map[string]any{"model_whitelist": []string{"*"}}
	require.False(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
}

func TestIsModelSupported_OpenAIAPIKeyRequiresExplicitCustomScope(t *testing.T) {
	account := &acct.Record{ID: 2, Platform: capability.PlatformOpenAI, Type: capability.AccountTypeAPIKey}
	require.False(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
	require.True(t, account.IsModelSupported("gpt-5.4", ModelDefaults(), ModelRules(account)))
	account.Credentials = map[string]any{"model_whitelist": []string{"*"}}
	require.True(t, account.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(account)))
}

func TestIsModelSupported_AnthropicDefaultsRejectForeignModel(t *testing.T) {
	anthropic := &acct.Record{ID: 3, Platform: capability.PlatformAnthropic, Type: capability.AccountTypeOAuth}
	require.True(t, anthropic.IsModelSupported("claude-sonnet-4-6", ModelDefaults(), ModelRules(anthropic)))
	require.False(t, anthropic.IsModelSupported("deepseek-v4", ModelDefaults(), ModelRules(anthropic)))
}
