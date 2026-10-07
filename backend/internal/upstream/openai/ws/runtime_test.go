package ws

import (
	"context"
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

// TestPrewarmUpdatesDuringDial 检查迟到连接、剩余拨号名额和忙连接对在线缩容的响应。
func TestPrewarmUpdatesDuringDial(t *testing.T) {
	for _, tc := range []struct {
		name                                                      string
		minimum, maximum, capacity, busy, otherCreating, wantIdle int
	}{
		{name: "zero_idle", capacity: 4},
		{name: "lower_idle", minimum: 1, maximum: 1, capacity: 4, wantIdle: 1},
		{name: "lower_target", minimum: 1, maximum: 4, capacity: 4, wantIdle: 1},
		{name: "no_idle_target", maximum: 4, capacity: 4},
		{name: "lower_capacity", minimum: 1, maximum: 1, capacity: 1, wantIdle: 1},
		{name: "busy_connections", capacity: 1, busy: 2},
		{name: "busy_with_idle_budget", maximum: 1, capacity: 4, busy: 2, wantIdle: 1},
		{name: "other_dial_reservation", capacity: 4, otherCreating: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initial := WSPoolOptions{MaxConnsPerProvider: 4, MinIdlePerProvider: 4, MaxIdlePerProvider: 4, DialTimeoutSeconds: 2}
			pool := NewWSConnPool(&initial)
			t.Cleanup(pool.Close)
			dialer := newOpenAIWSFirstDialBlockingCaptureDialer()
			pool.SetClientDialerForTest(dialer)
			var releaseOnce sync.Once
			releaseDial := func() { releaseOnce.Do(func() { close(dialer.releaseFirst) }) }
			t.Cleanup(releaseDial)
			req := WSAcquireRequest{Provider: &WSPoolProvider{ID: 9001, Type: "apikey"}, WSURL: "wss://example.test/responses"}
			ap := pool.getOrCreateProviderPool(9001)
			ap.lastAcquire = CloneWSAcquireRequestPtr(&req)
			var busy []*WSConnLease
			for range tc.busy {
				conn := NewWSConn(pool.nextConnID(9001), 9001, &openAIWSFakeConn{}, nil, nil, "")
				require.True(t, conn.tryAcquire())
				ap.conns[conn.id] = conn
				lease := &WSConnLease{pool: pool, ProviderID: 9001, Conn: conn}
				busy = append(busy, lease)
				t.Cleanup(lease.Release)
			}
			pool.ensureTargetIdleAsync(9001)
			select {
			case <-dialer.firstStarted:
			case <-time.After(time.Second):
				t.Fatal("预热拨号未开始")
			}
			ap.mu.Lock()
			ap.creating += tc.otherCreating
			ap.mu.Unlock()
			next := initial
			next.MinIdlePerProvider, next.MaxIdlePerProvider, next.MaxConnsPerProvider = tc.minimum, tc.maximum, tc.capacity
			require.NoError(t, pool.UpdateOptions(next))
			releaseDial()
			pool.prewarmWG.Wait()
			state, ok := pool.SnapshotProviderState(9001)
			require.True(t, ok)
			require.Equal(t, tc.wantIdle+tc.busy, state.Connections)
			require.Equal(t, tc.busy, state.LeasedConnections)
			require.Equal(t, 1, dialer.DialCount(), "需求降低后应停止本批剩余拨号")
			ap.mu.Lock()
			creating, active := ap.creating, ap.prewarmActive
			ap.creating -= tc.otherCreating
			ap.mu.Unlock()
			require.Equal(t, tc.otherCreating, creating, "应归还本批的剩余名额，保留其他请求的名额")
			require.False(t, active)
			dialer.mu.Lock()
			connections := append([]*openAIWSFakeConn(nil), dialer.connections...)
			dialer.mu.Unlock()
			require.Len(t, connections, 1)
			connections[0].mu.Lock()
			closed := connections[0].closed
			connections[0].mu.Unlock()
			require.Equal(t, tc.wantIdle == 0, closed)
			for _, lease := range busy {
				require.NoError(t, lease.WriteJSON(map[string]any{"type": "response.create"}, time.Second))
				lease.Release()
			}
			// 空闲数量为零时，新请求仍可按需建连，释放后立即回收。
			if tc.maximum == 0 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				lease, err := pool.Acquire(ctx, req)
				require.NoError(t, err)
				lease.Release()
				state, _ = pool.SnapshotProviderState(9001)
				require.Zero(t, state.Connections)
			}
		})
	}
}
