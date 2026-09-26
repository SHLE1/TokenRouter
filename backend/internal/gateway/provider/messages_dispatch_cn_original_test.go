//go:build unit

package provider

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/stretchr/testify/require"
)

// 分组显式映射统一作用于候选模型，实际账号仍须通过模型资格检查。
func TestResolveMessagesDispatchModelUsesExplicitGroupMapping(t *testing.T) {
	group := &routing.Group{MessagesDispatchModelConfig: routing.OpenAIMessagesDispatchModelConfig{SonnetMappedModel: "gpt-5.4"}}
	require.Equal(t, "gpt-5.4", ResolveMessagesDispatchModel(group, "claude-sonnet-4-5"))
	require.Empty(t, ResolveMessagesDispatchModel(group, "claude-opus-4-6"))
}
