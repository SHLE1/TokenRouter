package capability

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAntigravityRetiredSnapshot 验证旧快照中手动声明的协议不能恢复调用资格。
func TestAntigravityRetiredSnapshot(t *testing.T) {
	protocols := []ProtocolID{ProtocolAnthropicMessages, ProtocolGeminiGenerateContent, ProtocolOpenAIChatCompletions, ProtocolOpenAIResponses}
	for _, kind := range []string{ProviderTypeAPIKey, ProviderTypeUpstream, ProviderTypeSetupToken} {
		for _, source := range protocols {
			_, ok := ResolveRoute(ProviderProtocols{Platform: PlatformAntigravity, Type: kind, Enabled: protocols}, source, nil)
			require.False(t, ok, "%s %s", kind, source)
		}
		require.Empty(t, NativeProtocolOptions(PlatformAntigravity, kind, ""))
	}
	for _, source := range protocols {
		target, ok := ResolveRoute(ProviderProtocols{Platform: PlatformAntigravity, Type: ProviderTypeOAuth, Enabled: []ProtocolID{ProtocolGeminiGenerateContent}}, source, nil)
		require.True(t, ok)
		require.Equal(t, ProtocolGeminiGenerateContent, target)
	}
}
