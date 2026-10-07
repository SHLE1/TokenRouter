package ws

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/stretchr/testify/require"
)

// TestWSPoolUnchangedDestinationReusesConnection 检查拨号参数首尾空白的规范化。
func TestWSPoolUnchangedDestinationReusesConnection(t *testing.T) {
	pool, req := newWSReuseTestPool(1, 1)
	defer pool.Close()
	req.ProxyURL = "http://proxy.example:8080"
	first, err := pool.Acquire(context.Background(), req)
	require.NoError(t, err)
	first.Release()
	req.WSURL = " " + req.WSURL + " "
	req.ProxyURL = " " + req.ProxyURL + " "
	second, err := pool.Acquire(context.Background(), req)
	require.NoError(t, err)
	defer second.Release()
	require.Equal(t, first.ConnID(), second.ConnID())
	require.True(t, second.Reused())
}

// TestWSPoolDestinationCompatibility 检查地址和代理变更后的连接选择。
func TestWSPoolDestinationCompatibility(t *testing.T) {
	for _, destination := range []string{"url", "proxy"} {
		for _, preference := range []string{"ordinary", "preferred", "required"} {
			t.Run(destination+"/"+preference, func(t *testing.T) {
				pool, req := newWSReuseTestPool(1, 1)
				defer pool.Close()
				first, err := pool.Acquire(context.Background(), req)
				require.NoError(t, err)
				first.Release()
				if destination == "url" {
					req.WSURL = "wss://new.example/v1/responses"
				} else {
					req.ProxyURL = "http://proxy.example:8080"
				}
				if preference != "ordinary" {
					req.PreferredConnID = first.ConnID()
				}
				req.ForcePreferredConn = preference == "required"
				second, err := pool.Acquire(context.Background(), req)
				if second != nil {
					defer second.Release()
				}
				if req.ForcePreferredConn {
					require.ErrorIs(t, err, openai.ErrOpenAIWSPreferredConnUnavailable)
					return
				}
				require.NoError(t, err)
				require.NotEqual(t, first.ConnID(), second.ConnID())
				require.False(t, second.Reused())
			})
		}
	}
}

// TestWSPoolDestinationChangeKeepsActiveLease 检查目标变更时正在使用的连接。
func TestWSPoolDestinationChangeKeepsActiveLease(t *testing.T) {
	pool, req := newWSReuseTestPool(2, 2)
	defer pool.Close()
	first, err := pool.Acquire(context.Background(), req)
	require.NoError(t, err)
	defer first.Release()
	req.WSURL = "wss://new.example/v1/responses"
	second, err := pool.Acquire(context.Background(), req)
	require.NoError(t, err)
	defer second.Release()
	require.NotEqual(t, first.ConnID(), second.ConnID())
	require.NoError(t, first.WriteJSONContext(context.Background(), map[string]string{"type": "response.create"}))
}

// TestWSPoolPrewarmDropsChangedDestination 检查地址和代理变化后迟到的预热结果。
func TestWSPoolPrewarmDropsChangedDestination(t *testing.T) {
	for _, destination := range []string{"url", "proxy"} {
		t.Run(destination, func(t *testing.T) {
			pool, req := newWSReuseTestPool(2, 2)
			defer pool.Close()
			dialer := newOpenAIWSFirstDialBlockingCaptureDialer()
			pool.SetClientDialerForTest(dialer)
			var release sync.Once
			unblock := func() { release.Do(func() { close(dialer.releaseFirst) }) }
			defer unblock()
			ap := pool.getOrCreateProviderPool(req.Provider.ID)
			ap.lastAcquire = CloneWSAcquireRequestPtr(&req)
			options := *pool.nativeOptions()
			options.MinIdlePerProvider = 1
			require.NoError(t, pool.UpdateOptions(options))
			<-dialer.firstStarted
			if destination == "url" {
				req.WSURL = "wss://new.example/v1/responses"
			} else {
				req.ProxyURL = "http://proxy.example:8080"
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			lease, err := pool.Acquire(ctx, req)
			require.NoError(t, err)
			defer lease.Release()
			unblock()
			pool.prewarmWG.Wait()
			state, _ := pool.SnapshotProviderState(req.Provider.ID)
			require.Equal(t, 1, state.Connections)
			dialer.mu.Lock()
			connections := append([]*openAIWSFakeConn(nil), dialer.connections...)
			dialer.mu.Unlock()
			require.Len(t, connections, 2)
			for _, conn := range connections {
				conn.mu.Lock()
				closed := conn.closed
				conn.mu.Unlock()
				require.Equal(t, conn != lease.Conn.ws, closed)
			}
		})
	}
}

// TestWSPoolWaiterUsesAnyCompatibleConnection 检查另一个会话先结束时的连接交付。
func TestWSPoolWaiterUsesAnyCompatibleConnection(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(2, 2)
		defer pool.Close()
		first, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		defer first.Release()
		second, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		defer second.Release()
		first.Conn.lastUsedNano.Store(time.Now().Add(-time.Second).UnixNano())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer func() {
			cancel()
			synctest.Wait()
		}()
		result := acquireWSInBackground(pool, ctx, req)
		synctest.Wait()
		second.Release()
		synctest.Wait()
		select {
		case got := <-result:
			require.NoError(t, got.err)
			defer got.lease.Release()
			require.Equal(t, second.ConnID(), got.lease.ConnID())
		default:
			t.Fatal("兼容连接已归还，等待者仍未取得租约")
		}
	})
}

// TestWSPoolIdleLimitExcludesLeasedConnections 检查繁忙期间的空闲连接保留数量。
func TestWSPoolIdleLimitExcludesLeasedConnections(t *testing.T) {
	pool, req := newWSReuseTestPool(128, 12)
	defer pool.Close()
	leases := make([]*WSConnLease, 0, 22)
	for range 22 {
		lease, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		defer lease.Release()
		leases = append(leases, lease)
	}
	leases[20].Release()
	leases[21].Release()
	state, ok := pool.SnapshotProviderState(req.Provider.ID)
	require.True(t, ok)
	require.Equal(t, 20, state.LeasedConnections)
	require.Equal(t, 22, state.Connections)
}

// TestWSPoolShrinkRespectsBusyAndIdleLimits 检查缩容时忙连接、空闲数量和回收顺序。
func TestWSPoolShrinkRespectsBusyAndIdleLimits(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		capacity, idle, want int
	}{
		{name: "idle_limit", capacity: 4, idle: 2, want: 3},
		{name: "capacity_limit", capacity: 2, idle: 2, want: 2},
		{name: "zero_idle", capacity: 4, idle: 0, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool, req := newWSReuseTestPool(4, 4)
			defer pool.Close()
			leases := make([]*WSConnLease, 0, 4)
			for range 4 {
				lease, err := pool.Acquire(context.Background(), req)
				require.NoError(t, err)
				defer lease.Release()
				leases = append(leases, lease)
			}
			for i := 1; i < len(leases); i++ {
				leases[i].Release()
				leases[i].Conn.lastUsedNano.Store(time.Now().Add(time.Duration(i-4) * time.Second).UnixNano())
			}
			options := *pool.nativeOptions()
			options.MaxConnsPerProvider = tc.capacity
			options.MaxIdlePerProvider = tc.idle
			require.NoError(t, pool.UpdateOptions(options))
			state, _ := pool.SnapshotProviderState(req.Provider.ID)
			require.Equal(t, tc.want, state.Connections)
			require.Equal(t, 1, state.LeasedConnections)
			require.NoError(t, leases[0].WriteJSONContext(context.Background(), map[string]string{"type": "response.create"}))
			// 最久未使用的空闲连接先被回收。
			select {
			case <-leases[1].Conn.closedCh:
			default:
				t.Fatal("最久未使用的空闲连接仍在池中")
			}
			leases[0].Release()
			state, _ = pool.SnapshotProviderState(req.Provider.ID)
			require.Equal(t, min(tc.idle, tc.capacity), state.Connections)
		})
	}
}

// TestWSPoolZeroIdleHandsConnectionToWaiter 检查禁留空闲连接时的租约交接。
func TestWSPoolZeroIdleHandsConnectionToWaiter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(1, 0)
		defer pool.Close()
		first, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		defer first.Release()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer func() {
			cancel()
			synctest.Wait()
		}()
		result := acquireWSInBackground(pool, ctx, req)
		synctest.Wait()
		first.Release()
		synctest.Wait()
		got := <-result
		require.NoError(t, got.err)
		require.Equal(t, first.ConnID(), got.lease.ConnID())
		got.lease.Release()
		state, _ := pool.SnapshotProviderState(req.Provider.ID)
		require.Zero(t, state.Connections)
	})
}

// newWSReuseTestPool 使用本地替身拨号器，并让预热目标等于当前租约数量。
func newWSReuseTestPool(capacity, idle int) (*WSConnPool, WSAcquireRequest) {
	pool := NewWSConnPool(&WSPoolOptions{
		MaxConnsPerProvider:   capacity,
		MaxIdlePerProvider:    idle,
		PoolTargetUtilization: 1,
		QueueLimitPerConn:     2,
	})
	pool.SetClientDialerForTest(&openAIWSCountingDialer{})
	return pool, WSAcquireRequest{
		Provider: &WSPoolProvider{ID: 1, Type: "apikey"},
		WSURL:    "wss://example.com/v1/responses",
	}
}

// wsAcquireTestResult 将后台获取结果送回测试协程。
type wsAcquireTestResult struct {
	lease *WSConnLease
	err   error
}

// acquireWSInBackground 启动一次有取消预算的连接获取。
func acquireWSInBackground(pool *WSConnPool, ctx context.Context, req WSAcquireRequest) <-chan wsAcquireTestResult {
	result := make(chan wsAcquireTestResult, 1)
	go func() {
		lease, err := pool.Acquire(ctx, req)
		result <- wsAcquireTestResult{lease: lease, err: err}
	}()
	return result
}
