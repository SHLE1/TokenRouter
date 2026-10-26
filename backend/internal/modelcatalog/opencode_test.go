package modelcatalog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOpenCodeProtocolsSeparatesModesAndCopiesDefaults compares an exact shared
// model across the bundled service offerings. Mutating one result cannot alter
// subsequent route lookups or another service mode.
func TestOpenCodeProtocolsSeparatesModesAndCopiesDefaults(t *testing.T) {
	payg, goMode := OpenCodeProtocols("payg"), OpenCodeProtocols("go")
	require.Equal(t, "openai_chat_completions", payg["minimax-m3"])
	require.Equal(t, "anthropic_messages", goMode["minimax-m3"])
	require.Equal(t, "systemone", payg["jev-1.13"])
	payg["minimax-m3"] = "changed"
	require.Equal(t, "openai_chat_completions", OpenCodeProtocols("payg")["minimax-m3"])
	require.Empty(t, OpenCodeProtocols("unknown"))
}

// TestParseOpenCodeProtocolsUsesDeclaredSDK gives one provider default and
// several explicit model SDKs. Unknown encodings stay unavailable rather than
// being guessed from otherwise familiar model names.
func TestParseOpenCodeProtocolsUsesDeclaredSDK(t *testing.T) {
	body := []byte(`{"providers":{"opencode":{"npm":"@ai-sdk/openai-compatible","models":{"chat":{},"response":{"provider":{"npm":"@ai-sdk/openai"}},"messages":{"provider":{"npm":"@ai-sdk/anthropic"}},"google":{"provider":{"npm":"@ai-sdk/google"}},"decision":{"type":"decision"},"gpt-unknown":{"provider":{"npm":"unknown"}}}}}}`)
	routes := ParseOpenCodeProtocols(body)["payg"]
	require.Len(t, routes, 5)
	require.Equal(t, "gemini_generate_content", routes["google"])
	require.Equal(t, "systemone", routes["decision"])
	require.NotContains(t, routes, "gpt-unknown")
}
