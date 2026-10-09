package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
)

// TestQueueReopensAndKeepsUpdates 覆盖进程重启，以及批写期间请求继续更新的情形。
func TestQueueReopensAndKeepsUpdates(t *testing.T) {
	dir := t.TempDir()
	queue, err := NewQueue(dir)
	require.NoError(t, err)
	start := time.Now().UTC()
	record := telemetry.RequestRecord{RequestID: "request", UserID: 42, StartedAt: start, UpdatedAt: start, State: "running", Path: "/v1/responses"}
	require.NoError(t, queue.Put(record))
	_, err = NewQueue(dir)
	require.Error(t, err)
	require.NoError(t, queue.Close())
	reopened, err := NewQueue(dir)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reopened.Close()) })
	batch, err := reopened.Peek(128)
	require.NoError(t, err)
	require.Len(t, batch, 1)
	finished := start.Add(time.Second)
	require.NoError(t, reopened.Put(telemetry.RequestRecord{RequestID: record.RequestID, UpdatedAt: finished, FinishedAt: &finished, State: "completed", Aliases: []telemetry.RequestAlias{{Kind: "billing", Value: "client:request"}}}))
	require.NoError(t, reopened.Ack(batch))
	current, ok := reopened.Get(record.RequestID)
	require.True(t, ok)
	require.Equal(t, int64(42), current.UserID)
	require.Equal(t, record.Path, current.Path)
	require.Equal(t, "completed", current.State)
	require.NoError(t, reopened.Put(record))
	batch, err = reopened.Peek(128)
	require.NoError(t, err)
	require.Equal(t, "completed", batch[0].State)
	require.NoError(t, reopened.Ack(batch))
	pending, err := reopened.Pending()
	require.NoError(t, err)
	require.Zero(t, pending)
}

// BenchmarkQueuePut 记录本地持久写入的单次成本，供吞吐验收使用。
func BenchmarkQueuePut(b *testing.B) {
	queue, err := NewQueue(b.TempDir())
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, queue.Close()) })
	for b.Loop() {
		now := time.Now().UTC()
		require.NoError(b, queue.Put(telemetry.RequestRecord{RequestID: "benchmark", StartedAt: now, UpdatedAt: now, State: "running"}))
	}
}

// TestQueueRecoversIncompleteTail 保留已同步快照，并忽略中断写入留下的末帧。
func TestQueueRecoversIncompleteTail(t *testing.T) {
	dir := t.TempDir()
	queue, err := NewQueue(dir)
	require.NoError(t, err)
	now := time.Now().UTC()
	require.NoError(t, queue.Put(telemetry.RequestRecord{RequestID: "kept", StartedAt: now, UpdatedAt: now, State: "running"}))
	require.NoError(t, queue.Close())
	file, err := os.OpenFile(filepath.Join(dir, "requests.journal"), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = file.WriteString(`{"record":{"request_id":"partial`)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	reopened, err := NewQueue(dir)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reopened.Close()) })
	record, ok := reopened.Get("kept")
	require.True(t, ok)
	require.Equal(t, "running", record.State)
	_, ok = reopened.Get("partial")
	require.False(t, ok)
}

// BenchmarkQueueParallel 以并发请求检查批量同步的吞吐。
func BenchmarkQueueParallel(b *testing.B) {
	queue, err := NewQueue(b.TempDir())
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, queue.Close()) })
	var sequence atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		id := fmt.Sprint(sequence.Add(1))
		for pb.Next() {
			now := time.Now().UTC()
			if err := queue.Put(telemetry.RequestRecord{RequestID: id, StartedAt: now, UpdatedAt: now, State: "running"}); err != nil {
				b.Error(err)
				return
			}
		}
	})
}
