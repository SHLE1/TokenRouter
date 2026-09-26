package routing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 空分组配置不自动替换模型，显式配置也不读取供应商动态默认值。
func TestMessagesDispatchUsesOnlyExplicitGroupConfiguration(t *testing.T) {
	reads := 0
	options := MessagesDispatchOptions{
		NormalizeModel: strings.TrimSpace,
		CrossClientModel: func() string {
			reads++
			return "dynamic-model"
		},
	}
	require.Empty(t, ResolveMessagesDispatchModel(nil, "claude-sonnet", options))
	require.Empty(t, ResolveMessagesDispatchModel(&Group{}, "", options))
	require.Empty(t, ResolveMessagesDispatchModel(&Group{}, "gpt-5.5", options))
	require.Empty(t, ResolveMessagesDispatchModel(&Group{}, "claude-sonnet", options))
	require.Zero(t, reads)
	require.Equal(t, "explicit-model", ResolveMessagesDispatchModel(&Group{MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{SonnetMappedModel: "explicit-model"}}, "claude-sonnet", options))
	require.Zero(t, reads)
}
