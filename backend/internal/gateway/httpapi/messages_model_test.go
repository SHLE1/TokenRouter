package httpapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Messages 不再增加协议专用映射层，保留通用分组映射结果和既有协议型号规范化。
func TestMessagesAccountModelKeepsGroupMapping(t *testing.T) {
	require.Equal(t, "group-model", ResolveOpenAIMessagesAccountLayerModel("group-model"))
	require.Equal(t, "claude-sonnet-4-6", ResolveOpenAIMessagesAccountLayerModel("claude-sonnet-4-6"))
	require.Equal(t, "gpt-5.4", ResolveOpenAIMessagesAccountLayerModel("gpt-5.4-high"))
}
