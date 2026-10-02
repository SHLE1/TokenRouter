package scheduler

import "github.com/TokenFlux/TokenRouter/internal/scheduler/policy"

// ShouldEscapeSticky 使用本次请求的策略和共享反馈，先检查 TTFT。
func ShouldEscapeSticky(stats *RuntimeStats, providerID int64, cfg policy.StickyEscapeConfig) (reason string, errorRate float64, ttft float64, shouldEscape bool) {
	if !cfg.Enabled || stats == nil || providerID <= 0 {
		return "", 0, 0, false
	}
	errorRate, ttft, hasTTFT := stats.Snapshot(providerID)
	if hasTTFT && ttft > cfg.TtftMs {
		return "ttft", errorRate, ttft, true
	}
	if errorRate > cfg.ErrorRate {
		return "error_rate", errorRate, ttft, true
	}
	return "", errorRate, ttft, false
}
