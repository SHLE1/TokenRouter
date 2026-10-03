package provider

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/stretchr/testify/require"
)

// TestDetachedUpstreamReceivesInternalAbort 覆盖图片和流式上游的取消隔离入口。
func TestDetachedUpstreamReceivesInternalAbort(t *testing.T) {
	for _, stream := range []bool{false, true} {
		synctest.Test(t, func(t *testing.T) {
			client, cancelClient := context.WithCancel(context.Background())
			request, abort := apikey.WithRequestAbort(client)
			defer abort()
			var upstream context.Context
			var release context.CancelFunc
			if stream {
				upstream, release = DetachStreamUpstreamContext(request, true)
			} else {
				upstream, release = DetachUpstreamContext(request)
			}
			defer release()
			cancelClient()
			synctest.Wait()
			require.NoError(t, upstream.Err())
			abort()
			synctest.Wait()
			require.ErrorIs(t, upstream.Err(), context.Canceled)
		})
	}
}
