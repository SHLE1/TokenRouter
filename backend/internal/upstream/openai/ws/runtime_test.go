package ws

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// delayedCloseConn 用阻塞关闭握手检查配置更新期间的连接池访问。
type delayedCloseConn struct {
	openAIWSFakeConn
	entered chan struct{}
	release chan struct{}
}

func (c *delayedCloseConn) Close() error {
	close(c.entered)
	<-c.release
	return nil
}

func TestConnectionsShrinkDoesNotBlockPoolAccess(t *testing.T) {
	initial := &WSPoolOptions{MaxConnsPerProvider: 2, MaxIdlePerProvider: 2}
	pool := NewWSConnPool(initial)
	owner := NewOpenAIWSConnections(initial, nil, pool)
	t.Cleanup(owner.Close)
	slow := &delayedCloseConn{entered: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(slow.release) }) }
	t.Cleanup(unblock)
	provider := pool.getOrCreateProviderPool(1)
	provider.conns["old"] = NewWSConn("old", 1, slow, nil, nil, "")
	provider.conns["old"].lastUsedNano.Store(time.Now().Add(-time.Second).UnixNano())
	provider.conns["new"] = NewWSConn("new", 1, nil, nil, nil, "")
	updated := make(chan error, 1)
	go func() {
		updated <- owner.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: 1, MaxIdlePerProvider: 1})
	}()
	select {
	case <-slow.entered:
	case <-time.After(time.Second):
		t.Fatal("关闭握手没有开始")
	}
	available := make(chan *WSConnPool, 1)
	go func() { available <- owner.Pool() }()
	select {
	case actual := <-available:
		require.Same(t, pool, actual)
	case <-time.After(time.Second):
		t.Error("获取连接池被上游关闭握手阻塞")
	}
	unblock()
	require.NoError(t, <-updated)
}

func TestConnectionsConcurrentUpdatesKeepOwnerAndPoolConsistent(t *testing.T) {
	owner := NewOpenAIWSConnections(&WSPoolOptions{MaxConnsPerProvider: 4}, nil)
	t.Cleanup(owner.Close)
	pool := owner.Pool()
	var workers sync.WaitGroup
	for worker := range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := range 25 {
				if err := owner.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: 1 + worker*25 + i}); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	workers.Wait()
	require.Equal(t, *owner.Options, *pool.nativeOptions())
	owner.Close()
	require.Error(t, owner.UpdateOptions(WSPoolOptions{MaxConnsPerProvider: 8}))
	require.Same(t, pool, owner.Pool())
}

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
