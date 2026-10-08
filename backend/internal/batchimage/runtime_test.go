package batchimage

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestCleanupImmediateStop(t *testing.T) {
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	core := &Cleanup{Repo: &cleanupRepo{}}
	s := NewRuntime("batch image cleanup", true, core.Run)
	s.Start()
	s.Stop()
}

type cleanupRepo struct {
	BatchImageRepository
}

func (*cleanupRepo) ListBatchImageJobsDueForInputCleanup(ctx context.Context, _ time.Time, _ int) ([]*BatchImageJob, error) {
	return nil, ctx.Err()
}

func (*cleanupRepo) ListBatchImageJobsDueForOutputCleanup(ctx context.Context, _ time.Time, _ int) ([]*BatchImageJob, error) {
	return nil, ctx.Err()
}
