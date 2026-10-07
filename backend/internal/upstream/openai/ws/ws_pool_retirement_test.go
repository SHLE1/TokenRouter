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

// TestWSPoolExpiredConnectionIsRetiredBeforeHandoff 检查排队请求对过期连接的处理。
func TestWSPoolExpiredConnectionIsRetiredBeforeHandoff(t *testing.T) {
	for _, required := range []bool{false, true} {
		name := "ordinary"
		if required {
			name = "required"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				pool, req := newWSReuseTestPool(1, 1)
				defer pool.Close()
				first, err := pool.Acquire(context.Background(), req)
				require.NoError(t, err)
				defer first.Release()
				first.Conn.createdAtNano.Store(time.Now().Add(-WSConnMaxAge + time.Millisecond).UnixNano())
				if required {
					req.PreferredConnID = first.ConnID()
					req.ForcePreferredConn = true
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer func() {
					cancel()
					synctest.Wait()
				}()
				result := acquireWSInBackground(pool, ctx, req)
				synctest.Wait()
				_, waiters, _ := pool.ProviderPoolLoad(req.Provider.ID)
				require.Equal(t, 1, waiters)
				time.Sleep(2 * time.Millisecond)
				// 会话仍持有租约时可以完成当前工作。
				require.NoError(t, first.WriteJSONContext(ctx, map[string]string{"type": "response.create"}))
				first.Release()
				synctest.Wait()
				got := <-result
				if got.lease != nil {
					defer got.lease.Release()
				}
				if required {
					require.ErrorIs(t, got.err, openai.ErrOpenAIWSPreferredConnUnavailable)
				} else {
					require.NoError(t, got.err)
					require.NotEqual(t, first.ConnID(), got.lease.ConnID())
				}
			})
		})
	}
}

// TestWSPoolChecksAgeBetweenCleanupSweeps 检查连接在两次周期清理之间过期的情况。
func TestWSPoolChecksAgeBetweenCleanupSweeps(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(1, 1)
		defer pool.Close()
		first, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		first.Conn.createdAtNano.Store(time.Now().Add(-WSConnMaxAge + time.Millisecond).UnixNano())
		first.Release()
		time.Sleep(time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		second, err := pool.Acquire(ctx, req)
		require.NoError(t, err)
		defer second.Release()
		require.NotEqual(t, first.ConnID(), second.ConnID())
	})
}

// TestWSPoolCanceledAcquireDoesNotWaitForClose 检查取消返回与后台关闭握手的时序。
func TestWSPoolCanceledAcquireDoesNotWaitForClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(1, 0)
		pool.Start()
		defer pool.Close()
		slow := &delayedCloseConn{entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		unblock := func() { release.Do(func() { close(slow.release) }) }
		defer unblock()
		conn := NewWSConn("cancel-close", req.Provider.ID, slow, nil, nil, "")
		conn.compatibility = wsCompatibilityForRequest(req)
		require.True(t, conn.tryAcquire())
		ap := pool.getOrCreateProviderPool(req.Provider.ID)
		ap.conns[conn.id] = conn
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result := acquireWSInBackground(pool, ctx, req)
		synctest.Wait()
		ap.mu.Lock()
		conn.release()
		cancel()
		ap.mu.Unlock()
		synctest.Wait()
		select {
		case got := <-result:
			require.ErrorIs(t, got.err, context.Canceled)
		default:
			t.Fatal("取消返回被关闭握手阻塞")
		}
		select {
		case <-slow.entered:
		default:
			t.Fatal("后台清理没有开始关闭空闲连接")
		}
		_, waiters, conns := pool.ProviderPoolLoad(req.Provider.ID)
		require.Zero(t, waiters)
		require.Zero(t, conns)
		// 池关闭需要等清理任务完成，已经取消的请求可以先返回。
		stopped := make(chan struct{})
		go func() {
			pool.Close()
			close(stopped)
		}()
		synctest.Wait()
		select {
		case <-stopped:
			t.Fatal("池在后台清理结束前完成了关闭")
		default:
		}
		unblock()
		synctest.Wait()
		<-stopped
	})
}

// TestWSPoolCanceledDeliveryClosesInBackground 检查领取租约后、交付前发生取消的情况。
func TestWSPoolCanceledDeliveryClosesInBackground(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(1, 0)
		pool.Start()
		defer pool.Close()
		slow := &delayedCloseConn{entered: make(chan struct{}), release: make(chan struct{})}
		defer close(slow.release)
		conn := NewWSConn("undelivered", req.Provider.ID, slow, nil, nil, "")
		conn.compatibility = wsCompatibilityForRequest(req)
		require.True(t, conn.tryAcquire())
		ap := pool.getOrCreateProviderPool(req.Provider.ID)
		ap.conns[conn.id] = conn
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := make(chan wsAcquireTestResult, 1)
		go func() {
			lease, err := pool.deliverLease(ctx, req, 0, conn, true, 0)
			result <- wsAcquireTestResult{lease: lease, err: err}
		}()
		synctest.Wait()
		select {
		case got := <-result:
			require.Nil(t, got.lease)
			require.ErrorIs(t, got.err, context.Canceled)
		default:
			t.Fatal("取消交付时同步等待了关闭握手")
		}
		select {
		case <-slow.entered:
		default:
			t.Fatal("未交付的连接没有进入后台回收")
		}
	})
}

// TestWSPoolShutdownCloseDoesNotBlockWaiterExit 检查池关闭期间的等待者退出。
func TestWSPoolShutdownCloseDoesNotBlockWaiterExit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool, req := newWSReuseTestPool(2, 2)
		defer pool.Close()
		first, err := pool.Acquire(context.Background(), req)
		require.NoError(t, err)
		defer first.Release()
		slow := &delayedCloseConn{entered: make(chan struct{}), release: make(chan struct{})}
		var release sync.Once
		unblock := func() { release.Do(func() { close(slow.release) }) }
		defer unblock()
		idle := NewWSConn("shutdown-idle", req.Provider.ID, slow, nil, nil, "")
		idle.compatibility = wsCompatibilityForRequest(req)
		ap := pool.getOrCreateProviderPool(req.Provider.ID)
		ap.conns[idle.id] = idle
		req.PreferredConnID = first.ConnID()
		req.ForcePreferredConn = true
		result := acquireWSInBackground(pool, context.Background(), req)
		synctest.Wait()
		stopped := make(chan struct{})
		go func() {
			pool.Close()
			close(stopped)
		}()
		synctest.Wait()
		select {
		case <-slow.entered:
		default:
			t.Fatal("池关闭没有开始回收空闲连接")
		}
		select {
		case got := <-result:
			require.ErrorIs(t, got.err, errOpenAIWSConnClosed)
		default:
			t.Fatal("等待者退出被池关闭握手阻塞")
		}
		unblock()
		synctest.Wait()
		<-stopped
	})
}

// TestWSPoolHandoffPrecedesIdleClose 检查交付租约与关闭多余空闲连接的时序。
func TestWSPoolHandoffPrecedesIdleClose(t *testing.T) {
	for _, required := range []bool{false, true} {
		name := "ordinary"
		if required {
			name = "required"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				pool, req := newWSReuseTestPool(2, 0)
				pool.Start()
				defer pool.Close()
				slow := &delayedCloseConn{entered: make(chan struct{}), release: make(chan struct{})}
				unblock := sync.OnceFunc(func() { close(slow.release) })
				defer unblock()
				first := NewWSConn("available", req.Provider.ID, &openAIWSFakeConn{}, nil, nil, "")
				second := NewWSConn("slow-idle", req.Provider.ID, slow, nil, nil, "")
				first.compatibility = wsCompatibilityForRequest(req)
				second.compatibility = first.compatibility
				require.True(t, first.tryAcquire())
				require.True(t, second.tryAcquire())
				ap := pool.getOrCreateProviderPool(req.Provider.ID)
				ap.conns[first.id] = first
				ap.conns[second.id] = second
				if required {
					req.PreferredConnID = first.id
					req.ForcePreferredConn = true
				}
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				result := acquireWSInBackground(pool, ctx, req)
				synctest.Wait()
				// 两个租约已归还，等待者和归还方接着争取池的访问锁。
				ap.mu.Lock()
				first.release()
				second.release()
				first.lastUsedNano.Store(time.Now().Add(-time.Millisecond).UnixNano())
				ap.signalChangedLocked()
				ap.mu.Unlock()
				synctest.Wait()
				var lease *WSConnLease
				select {
				case got := <-result:
					require.NoError(t, got.err)
					lease = got.lease
					defer lease.Release()
					require.Equal(t, first.id, lease.ConnID())
				default:
					t.Fatal("租约交付被另一条空闲连接的关闭握手阻塞")
				}
				require.NoError(t, ctx.Err())
				select {
				case <-slow.entered:
				default:
					t.Fatal("后台任务没有回收多余空闲连接")
				}
				require.NoError(t, lease.WriteJSONContext(ctx, map[string]string{"type": "response.create"}))
				state, _ := pool.SnapshotProviderState(req.Provider.ID)
				require.Equal(t, 1, state.Connections)
				require.Equal(t, 1, state.LeasedConnections)
				unblock()
				synctest.Wait()
				lease.Release()
				state, _ = pool.SnapshotProviderState(req.Provider.ID)
				require.Zero(t, state.Connections)
			})
		})
	}
}
