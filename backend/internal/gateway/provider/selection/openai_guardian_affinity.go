package selection

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
)

// resolveOpenAIGuardianParentProviderID 根据请求状态中的父会话散列查询缓存。
func (s *Compatible) resolveOpenAIGuardianParentProviderID(ctx context.Context, groupID *int64) int64 {
	if s == nil || s.cache == nil {
		return 0
	}
	affinity, ok := requeststate.GuardianParentAffinityFromContext(ctx)
	if !ok {
		return 0
	}
	lookupCtx := requeststate.WithOpenAILegacySessionHash(ctx, affinity.LegacySessionHash)
	providerID, err := s.getStickySessionProviderID(lookupCtx, groupID, affinity.CurrentSessionHash)
	if err != nil || providerID <= 0 {
		return 0
	}
	return providerID
}
