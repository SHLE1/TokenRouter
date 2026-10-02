package provider

import (
	"context"
	"reflect"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// matches 按当前记录身份比较观测，管理员更新身份后拒绝此前的结果。
func (r *qoderObservationIdentityRepo) matches(value provider.UsageObservationVersion) bool {
	current := r.current
	return current.ID == value.ID && current.Platform == value.Platform && current.Type == value.Type && current.Status == value.Status && reflect.DeepEqual(current.Credentials, value.Credentials) && reflect.DeepEqual(current.ProxyID, value.ProxyID)
}

func (r *qoderObservationIdentityRepo) UpdateUsageExtraIfUnchanged(ctx context.Context, value provider.UsageObservationVersion, updates map[string]any) (bool, error) {
	if !r.matches(value) {
		return false, nil
	}
	return true, r.UpdateExtra(ctx, value.ID, updates)
}

func (r *qoderObservationIdentityRepo) SetUsageRateLimitIfUnchanged(ctx context.Context, value provider.UsageObservationVersion, reset time.Time) (bool, error) {
	if !r.matches(value) {
		return false, nil
	}
	return true, r.SetRateLimited(ctx, value.ID, reset)
}

func (r *qoderObservationIdentityRepo) ClearUsageRateLimitIfUnchanged(ctx context.Context, value provider.UsageObservationVersion) (bool, error) {
	if !r.matches(value) {
		return false, nil
	}
	return true, r.ClearRateLimit(ctx, value.ID)
}
