package opencode

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestApplyHeadersRetainsClientIdentity replaces conflicting credentials while
// carrying the actual client agent and stable native session. Missing caller
// sessions stay absent rather than becoming random request identities.
func TestApplyHeadersRetainsClientIdentity(t *testing.T) {
	headers := http.Header{"authorization": []string{"wrong"}, "x-api-key": []string{"wrong"}, "x-goog-api-key": []string{"wrong"}}
	incoming := http.Header{"User-Agent": []string{"my-agent/2"}, "Session-Id": []string{"conversation"}}
	ApplyHeaders(headers, "secret", incoming.Get)
	require.Equal(t, "Bearer secret", headers.Get("Authorization"))
	require.Equal(t, "my-agent/2", headers.Get("User-Agent"))
	require.Equal(t, "conversation", headers.Get("X-OpenCode-Session"))
	require.Empty(t, headers.Get("x-api-key"))
	require.Empty(t, headers.Get("x-goog-api-key"))
	empty := make(http.Header)
	ApplyHeaders(empty, "secret", nil)
	require.Empty(t, empty.Get("X-OpenCode-Session"))
	require.Equal(t, "TokenRouter", empty.Get("User-Agent"))
}
