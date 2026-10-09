package requestlog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry"
)

type testQueue struct{ records []telemetry.RequestRecord }

type testRepository struct {
	err   error
	saved []telemetry.RequestRecord
}

func (q *testQueue) Put(record telemetry.RequestRecord) error {
	q.records = append(q.records, record)
	return nil
}

func (q *testQueue) Peek(int) ([]telemetry.RequestRecord, error) {
	return append([]telemetry.RequestRecord(nil), q.records...), nil
}

func (q *testQueue) Ack(records []telemetry.RequestRecord) error {
	q.records = q.records[len(records):]
	return nil
}
func (q *testQueue) Pending() (int, error) { return len(q.records), nil }

func (r *testRepository) Save(_ context.Context, records []telemetry.RequestRecord) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, records...)
	return nil
}

func (r *testRepository) Find(context.Context, string, int64, bool) ([]Detail, error) {
	return nil, r.err
}
func (r *testRepository) Cleanup(context.Context, time.Time) error { return nil }

// TestFlushRetriesDatabaseFailure 确认数据库失败时保留待写数据，恢复后才确认队列。
func TestFlushRetriesDatabaseFailure(t *testing.T) {
	queue := &testQueue{}
	repo := &testRepository{err: errors.New("database unavailable")}
	service := NewService(repo, queue, 30, nil)
	service.Observe(telemetry.RequestRecord{RequestID: "id", State: "failed"})
	require.Error(t, service.Flush(context.Background()))
	require.Len(t, queue.records, 1)
	repo.err = nil
	require.NoError(t, service.Flush(context.Background()))
	require.Empty(t, queue.records)
	require.Len(t, repo.saved, 1)
}
