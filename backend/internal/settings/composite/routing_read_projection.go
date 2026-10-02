package composite

import "github.com/TokenFlux/TokenRouter/internal/routing"

// ApplyRoutingAdminReadSettings 将路由读取结果写入综合快照。
func (s *Snapshot) ApplyRoutingAdminReadSettings(value *routing.AdminReadSettings) {
	s.EnableModelFallback = value.EnableModelFallback
	s.FallbackModelAnthropic = value.FallbackModelAnthropic
	s.FallbackModelAntigravity = value.FallbackModelAntigravity
	s.FallbackModelGemini = value.FallbackModelGemini
	s.FallbackModelOpenAI = value.FallbackModelOpenAI
	s.MarketplaceAvailabilityBucketMinutes = value.MarketplaceAvailabilityBucketMinutes
	s.MarketplaceAvailabilityWindowDays = value.MarketplaceAvailabilityWindowDays
}
