package httpx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSnapshotContextCacheKeepsETagAndIsolation 上下文取消支持保留快照 ETag 和返回值隔离。
func TestSnapshotContextCacheKeepsETagAndIsolation(t *testing.T) {
	cache := NewSnapshotCache(time.Minute)
	first, hit, err := cache.GetOrLoadContext(context.Background(), "snapshot", func(context.Context) (any, error) { return map[string]int{"requests": 17}, nil })
	require.NoError(t, err)
	require.False(t, hit)
	require.NotEmpty(t, first.ETag)
	firstPayload, typed := first.Payload.(map[string]int)
	require.True(t, typed)
	firstPayload["requests"] = 99
	second, hit, err := cache.GetOrLoadContext(context.Background(), "snapshot", func(context.Context) (any, error) { t.Fatal("缓存命中不应回源"); return nil, nil })
	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, first.ETag, second.ETag)
	secondPayload, typed := second.Payload.(map[string]int)
	require.True(t, typed)
	require.Equal(t, 17, secondPayload["requests"])
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = cache.GetOrLoadContext(canceled, "snapshot", func(context.Context) (any, error) { return nil, nil })
	require.ErrorIs(t, err, context.Canceled)
}
