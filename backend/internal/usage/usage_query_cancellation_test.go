package usage

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// statsCancellationRepo 将查询停在读取阶段，便于交错两个 HTTP 请求的取消。
type statsCancellationRepo struct {
	UsageLogRepository
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

// GetStatsWithFilters 等待测试释放查询或底层上下文取消。
func (r *statsCancellationRepo) GetStatsWithFilters(ctx context.Context, _ UsageLogFilters) (*UsageStats, error) {
	if r.calls.Add(1) == 1 {
		close(r.started)
	}
	select {
	case <-r.release:
		return &UsageStats{TotalRequests: 17}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// joinedStatsContext 在等待者进入取消选择时通知测试。
type joinedStatsContext struct {
	context.Context
	joined chan struct{}
	once   sync.Once
}

// Done 在等待登记完成后通知主测试协程。
func (c *joinedStatsContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.joined) })
	return c.Context.Done()
}

// TestStatsCanceledLeaderKeepsLiveWaiter 复现首请求退出时同筛选的另一个请求仍在等待。
func TestStatsCanceledLeaderKeepsLiveWaiter(t *testing.T) {
	repo := &statsCancellationRepo{started: make(chan struct{}), release: make(chan struct{})}
	service := NewUsageService(repo)
	filters := UsageLogFilters{EndpointSource: "upstream"}
	first, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstDone := make(chan error, 1)
	go func() { _, _, err := service.GetStatsCached(first, filters); firstDone <- err }()
	<-repo.started
	live := &joinedStatsContext{Context: context.Background(), joined: make(chan struct{})}
	nextDone := make(chan error, 1)
	nextValue := make(chan *UsageStats, 1)
	go func() { value, _, err := service.GetStatsCached(live, filters); nextValue <- value; nextDone <- err }()
	<-live.joined
	cancel()
	require.ErrorIs(t, <-firstDone, context.Canceled)
	close(repo.release)
	require.NoError(t, <-nextDone)
	require.EqualValues(t, 17, (<-nextValue).TotalRequests)
	require.EqualValues(t, 1, repo.calls.Load())
}
