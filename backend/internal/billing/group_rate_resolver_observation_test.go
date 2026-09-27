package billing

func GroupRateCacheStats() (int64, int64, int64, int64, int64) {
	return groupRateMetrics.Hit.Load(), groupRateMetrics.Miss.Load(), groupRateMetrics.Load.Load(), groupRateMetrics.Shared.Load(), groupRateMetrics.Fallback.Load()
}
