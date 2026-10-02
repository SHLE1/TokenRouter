package provider

import (
	"context"
)

type GroupReference struct {
	RequireOAuthOnly bool
	ID               int64
	Name             string
}
type AdminGroups interface {
	ActiveGroups(context.Context, string) ([]GroupReference, error)
	ValidateGroups(context.Context, []int64) error
	GetGroup(context.Context, int64) (*GroupReference, error)
}
type CreateCredentialHooks struct {
	// 平台接口按 Qoder 站点切换和 PAT 校验的顺序执行。
	Site         func(*Record) (string, error)
	ValidateEdit func(context.Context, *Record, bool) error
	Prepare      func(*Record)
	Validate     func(context.Context, *Record) error
}

// DuplicateStore 拥有原提供商与关联一次提交，不执行平台交换。
type DuplicateStore interface {
	CreateWithProviderGroups(context.Context, *Record, []GroupMembership) error
}

// ShadowProxyStore 按顺序将母提供商的代理配置同步给影子。
type ShadowProxyStore interface {
	ListShadowsByParent(context.Context, int64) ([]*Record, error)
	Update(context.Context, *Record) error
}
