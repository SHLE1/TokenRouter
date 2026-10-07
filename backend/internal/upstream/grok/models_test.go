package grok

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestExplicitGrokIDsRemainUnchanged 验证文本、媒体和供应商限定名均保留原值。
func TestExplicitGrokIDsRemainUnchanged(t *testing.T) {
	for _, model := range []string{"grok", "grok-latest", "grok-4.6-latest", "grok-build-latest", "grok-4.20-multi-agent", "xai/grok-4.6", "grok-imagine-video-1.5-preview"} {
		require.Equal(t, model, NormalizeModelID(model))
		require.Equal(t, model, ResolveGrokTextResponsesModelID(model, "custom-default"))
		require.Equal(t, model, CanonicalImagineVideoModel(model))
	}
	require.Equal(t, "custom-default", ResolveGrokTextResponsesModelID("", "custom-default"))
}
