package scheduler

import "context"

// requestLeaseKey 标识 context 中管理当前请求资源的 Lease。
type requestLeaseKey struct{}

// WithRequestLease 将已经取得的用户租约传递给后续提供商尝试，执行层继续决定释放时机。
func WithRequestLease(ctx context.Context, lease *Lease) context.Context {
	return context.WithValue(ctx, requestLeaseKey{}, lease)
}

func RequestLease(ctx context.Context) *Lease {
	if ctx == nil {
		return nil
	}
	lease, _ := ctx.Value(requestLeaseKey{}).(*Lease)
	return lease
}

// ownRequestResource 登记请求异常返回时的清理函数，attempt 完成和请求清理共用一次释放。
func ownRequestResource(ctx context.Context, release func()) {
	if owner := RequestLease(ctx); owner != nil {
		owner.Own(release)
	}
}
