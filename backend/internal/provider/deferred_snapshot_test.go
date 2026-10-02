package provider

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestDeferredFailureKeepsNewerPendingActivity 检查旧批次写回失败后，待写入队列中较新的活动时间保持不变。
func TestDeferredFailureKeepsNewerPendingActivity(t *testing.T) {
	repo := &deferredDrainRepository{entered: make(chan struct{}), release: make(chan struct{}), err: errors.New("forced write failure")}
	wheel, err := NewTimingWheelService()
	require.NoError(t, err)
	svc := NewDeferredService(repo, wheel, DeferredOptions{Interval: time.Hour})
	svc.lastUsedUpdates.Store(int64(1), time.Unix(1, 0))
	flushed := make(chan error, 1)
	go func() { flushed <- svc.flushLastUsedErr() }()
	<-repo.entered
	svc.ScheduleLastUsedUpdate(1)
	expected, ok := svc.lastUsedUpdates.Load(int64(1))
	require.True(t, ok)
	close(repo.release)
	require.Error(t, <-flushed)
	actual, ok := svc.lastUsedUpdates.Load(int64(1))
	require.True(t, ok)
	require.Equal(t, expected, actual)
}
