package composite

import "github.com/TokenFlux/TokenRouter/internal/usage"

// ApplyUsageAdminReadSettings 将用量读取结果写入综合快照。
func (s *Snapshot) ApplyUsageAdminReadSettings(value *usage.AdminReadSettings) {
	s.AllowUserViewErrorRequests = value.AllowUserViewErrorRequests
	s.UsageRankingEnabled = value.UsageRankingEnabled
	s.UsageRankingLimit = value.UsageRankingLimit
	s.UsageRankingShowActualCost = value.UsageRankingShowActualCost
	s.UsageRankingShowRequests = value.UsageRankingShowRequests
	s.UsageRankingShowTotalTokens = value.UsageRankingShowTotalTokens
	s.UsageRankingSortBy = value.UsageRankingSortBy
}
