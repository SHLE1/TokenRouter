package httpapi

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/apikey"

	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/stretchr/testify/require"
)

// TestOpenAICompatibleRequestPlatformStaysUnspecifiedBeforeSelection 验证未选提供商时平台为空，候选可来自不同平台。
func TestOpenAICompatibleRequestPlatformStaysUnspecifiedBeforeSelection(t *testing.T) {
	require.Empty(t, OpenAICompatibleRequestPlatform(nil))
	require.Empty(t, OpenAICompatibleRequestPlatform(&apikey.APIKey{Group: &routing.Group{}}))
}
