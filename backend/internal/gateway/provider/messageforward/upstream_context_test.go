package messageforward

import (
	"context"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/pkg/requestcontext"
	"github.com/stretchr/testify/require"
)

// TestMessagesStreamAbort 验证流式 Messages 的构造后释放、断连和内部终止。
func TestMessagesStreamAbort(t *testing.T) {
	client, cancelClient := context.WithCancel(context.Background())
	request, abort := requestcontext.WithAbort(client)
	defer abort()
	upstream, release := detachedStreamContext(request, true)
	release()
	require.NoError(t, upstream.Err())
	cancelClient()
	require.NoError(t, upstream.Err())
	abort()
	require.ErrorIs(t, upstream.Err(), context.Canceled)
}
