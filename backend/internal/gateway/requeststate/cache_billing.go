package requeststate

import "context"

// cacheBillingKey 标识当前尝试的强制缓存计费状态。
type cacheBillingKey struct{}

// IsForceCacheBilling 读取强制缓存计费标记，缺失或类型错误时返回 false。
func IsForceCacheBilling(ctx context.Context) bool {
	value, _ := ctx.Value(cacheBillingKey{}).(bool)
	return value
}

// WithForceCacheBilling 派生尝试 context，不改写父请求或任何资金事实。
func WithForceCacheBilling(ctx context.Context) context.Context {
	return context.WithValue(ctx, cacheBillingKey{}, true)
}
