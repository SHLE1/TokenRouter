package ws

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolShrinkKeepsActiveLeaseAndRejectsUpdatesAfterClose(t *testing.T) {
	pool := NewWSConnPool(&WSPoolOptions{MaxConnsPerProvider: 4, MinIdlePerProvider: 0, MaxIdlePerProvider: 4})
	provider := pool.getOrCreateProviderPool(1)
	busy := NewWSConn("busy", 1, nil, nil, nil, "")
	idle := NewWSConn("idle", 1, nil, nil, nil, "")
	require.True(t, busy.tryAcquire())
	provider.conns[busy.id], provider.conns[idle.id] = busy, idle
	require.NoError(t, pool.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: 1, MinIdlePerProvider: 0, MaxIdlePerProvider: 0}))
	state, ok := pool.SnapshotProviderState(1)
	require.True(t, ok)
	require.Equal(t, 1, state.LeasedConnections)
	require.Equal(t, 1, state.Connections)
	(&WSConnLease{pool: pool, ProviderID: 1, Conn: busy}).Release()
	state, _ = pool.SnapshotProviderState(1)
	require.Zero(t, state.Connections)
	pool.Close()
	require.Error(t, pool.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: 2}))
}

func TestPoolConcurrentOptionsPublication(t *testing.T) {
	pool := NewWSConnPool(&WSPoolOptions{MaxConnsPerProvider: 4})
	defer pool.Close()
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 100; i++ {
				_ = pool.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: i + 1})
				_ = pool.nativeOptions().MaxConnsHardCap()
			}
		}()
	}
	workers.Wait()
}
