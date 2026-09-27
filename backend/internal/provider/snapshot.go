package provider

import (
	"slices"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// ProviderSnapshot 是候选判断需要的身份、资格与运行投影，不携带凭据或管理 Extra。
// 模型重写及平台凭据读取仍由各自显式入口提供，不允许从快照反查完整记录。
type ProviderSnapshot struct {
	// ModelPolicy 仅在实际模型匹配点装配，不随协议预检提前读取动态默认值。
	ModelPolicy      ModelRoutingSnapshot `json:"-"`
	ID               int64
	ParentProviderID *int64
	Platform         string
	Type             string
	AuthMode         string
	EnabledProtocols []capability.ProtocolID
	Status           string
	Schedulable      bool
	Concurrency      int
	Priority         int
	ExpiresAt        *time.Time
}

// RoutingSnapshot 返回请求独立副本；协议缺省解析仍由提供商配置规则唯一负责。
func (r *Record) RoutingSnapshot() ProviderSnapshot {
	if r == nil {
		return ProviderSnapshot{}
	}
	snapshot := ProviderSnapshot{
		ID: r.ID, Platform: r.Platform, Type: r.Type, AuthMode: ProtocolAuthMode(r),
		EnabledProtocols: slices.Clone(r.UpstreamProtocols()), Status: r.Status, Schedulable: r.Schedulable,
		Concurrency: r.Concurrency, Priority: r.Priority,
	}
	if r.ParentProviderID != nil {
		id := *r.ParentProviderID
		snapshot.ParentProviderID = &id
	}
	if r.ExpiresAt != nil {
		at := *r.ExpiresAt
		snapshot.ExpiresAt = &at
	}
	return snapshot
}

// Protocols 仅把已经解析好的能力传给纯目录，不暴露提供商存储结构。
func (s ProviderSnapshot) Protocols() capability.ProviderProtocols {
	return capability.ProviderProtocols{Platform: s.Platform, Type: s.Type, AuthMode: s.AuthMode, Enabled: slices.Clone(s.EnabledProtocols)}
}
