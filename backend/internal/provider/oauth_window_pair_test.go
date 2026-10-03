package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// windowPairReader 记录查询次数并提供失败和取消场景。
type windowPairReader struct {
	LocalUsageStats
	pairs    int
	singles  int
	starts   []time.Time
	failPair bool
	cancel   context.CancelFunc
}

// GetProviderWindowStatsPair 记录合并调用，并模拟可恢复错误或请求取消。
func (r *windowPairReader) GetProviderWindowStatsPair(_ context.Context, _ int64, a, b time.Time) (*WindowStats, *WindowStats, error) {
	r.pairs++
	r.starts = []time.Time{a, b}
	if r.cancel != nil {
		r.cancel()
		return nil, nil, context.Canceled
	}
	if r.failPair {
		return nil, nil, errors.New("pair unavailable")
	}
	value := &WindowStats{Requests: 7}
	return value, value, nil
}

// GetProviderWindowStats 模拟两个窗口中一项成功、另一项失败。
func (r *windowPairReader) GetProviderWindowStats(context.Context, int64, time.Time) (*WindowStats, error) {
	r.singles++
	if r.singles == 2 {
		return nil, errors.New("second unavailable")
	}
	return &WindowStats{Requests: 1}, nil
}

// TestOpenAIUsesOneWindowQuery 核对供应商重置时间和两个窗口结果的独立性。
func TestOpenAIUsesOneWindowQuery(t *testing.T) {
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	reset5 := now.Add(time.Hour)
	reset7 := now.Add(24 * time.Hour)
	reader := &windowPairReader{}
	stats := NewLocalUsageStatistics(reader, NewOAuthUsageCache(), LocalUsageStatisticsOptions{Now: func() time.Time { return now }})
	service := NewOAuthUsageService(nil, nil, stats, OAuthUsageOptions{Now: func() time.Time { return now }})
	value, err := service.GetOpenAIUsage(context.Background(), &Record{ID: 1, Platform: PlatformOpenAI, Type: ProviderTypeOAuth, Extra: map[string]any{
		"codex_5h_used_percent": 10.0, "codex_7d_used_percent": 20.0,
		"codex_5h_reset_at": reset5.Format(time.RFC3339), "codex_7d_reset_at": reset7.Format(time.RFC3339),
	}}, false)
	require.NoError(t, err)
	require.Equal(t, 1, reader.pairs)
	require.Zero(t, reader.singles)
	require.Equal(t, []time.Time{reset5.Add(-5 * time.Hour), reset7.Add(-7 * 24 * time.Hour)}, reader.starts)
	value.FiveHour.WindowStats.Requests = 99
	require.EqualValues(t, 7, value.SevenDay.WindowStats.Requests)
}

// TestWindowPairFailurePreservesPartialStats 合并失败时保留可读取窗口，取消时终止回退。
func TestWindowPairFailurePreservesPartialStats(t *testing.T) {
	reader := &windowPairReader{failPair: true}
	stats := NewLocalUsageStatistics(reader, nil, LocalUsageStatisticsOptions{})
	a, b := stats.GetWindowPair(context.Background(), 1, time.Now().Add(-time.Hour), time.Now())
	require.EqualValues(t, 1, a.Requests)
	require.Nil(t, b)
	require.Equal(t, 2, reader.singles)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader = &windowPairReader{cancel: cancel}
	stats = NewLocalUsageStatistics(reader, nil, LocalUsageStatisticsOptions{})
	a, b = stats.GetWindowPair(ctx, 1, time.Now(), time.Now())
	require.Nil(t, a)
	require.Nil(t, b)
	require.Zero(t, reader.singles)
}
