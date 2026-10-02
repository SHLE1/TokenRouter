package provider

import (
	"context"
	"reflect"

	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// ClearRefreshCooldownIfUnchanged 模拟交错执行的条件更新，PostgreSQL 集成测试覆盖行锁和 JSONB 比较。
func (r *refreshSuccessCooldownRepo) ClearRefreshCooldownIfUnchanged(ctx context.Context, v provider.RefreshCooldownVersion) (bool, error) {
	if !reflect.DeepEqual(provider.ObserveRefreshCooldown(r.current), v) {
		return false, nil
	}
	return true, r.ClearTempUnschedulable(ctx, v.ID)
}

func (r *tokenRefreshCandidateRepo) ClearRefreshCooldownIfUnchanged(_ context.Context, v provider.RefreshCooldownVersion) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.providers {
		if r.providers[i].ID == v.ID {
			if !reflect.DeepEqual(provider.ObserveRefreshCooldown(&r.providers[i]), v) {
				return false, nil
			}
			r.clearTempCalls++
			r.providers[i].TempUnschedulableUntil = nil
			r.providers[i].TempUnschedulableReason = ""
			return true, nil
		}
	}
	return false, nil
}
