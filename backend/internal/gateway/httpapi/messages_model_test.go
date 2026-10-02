package httpapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMessagesProviderModelKeepsGroupMapping 验证 Messages 使用通用分组映射结果并规范化协议型号。
func TestMessagesProviderModelKeepsGroupMapping(t *testing.T) {
	require.Equal(t, "group-model", ResolveOpenAIMessagesProviderLayerModel("group-model"))
	require.Equal(t, "claude-sonnet-4-6", ResolveOpenAIMessagesProviderLayerModel("claude-sonnet-4-6"))
	require.Equal(t, "gpt-5.4-high", ResolveOpenAIMessagesProviderLayerModel("gpt-5.4-high"))
}
