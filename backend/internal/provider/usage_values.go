package provider

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/upstream/usageview"
)

// UpstreamUsageQueryConfig 是提供商 Extra 中公开给管理员的非敏感查询配置。
type UpstreamUsageQueryConfig struct {
	Enabled bool   `json:"enabled"`
	Adapter string `json:"adapter"`
	BaseURL string `json:"base_url,omitempty"`
}

type UpstreamUsageAmount = usageview.UpstreamUsageAmount

type UpstreamUsageBalanceEntry = usageview.UpstreamUsageBalanceEntry

type UpstreamUsageLimit = usageview.UpstreamUsageLimit

type UpstreamUsageSubscription = usageview.UpstreamUsageSubscription

type UpstreamUsageInfo = usageview.UpstreamUsageInfo

// UpstreamUsageQueryResult 是管理员查询接口的成功响应。
type UpstreamUsageQueryResult struct {
	ProviderID   int64                       `json:"provider_id"`
	Adapter      string                      `json:"adapter"`
	ObservedAt   time.Time                   `json:"observed_at"`
	Provider     string                      `json:"provider,omitempty"`
	Mode         string                      `json:"mode,omitempty"`
	Unit         string                      `json:"unit,omitempty"`
	Balance      *UpstreamUsageAmount        `json:"balance,omitempty"`
	Balances     []UpstreamUsageBalanceEntry `json:"balances,omitempty"`
	Available    *bool                       `json:"available,omitempty"`
	Limits       []UpstreamUsageLimit        `json:"limits,omitempty"`
	Subscription *UpstreamUsageSubscription  `json:"subscription,omitempty"`
	ExpiresAt    *time.Time                  `json:"expires_at,omitempty"`
	// Usage 保存服务内部使用的归一化结果。
	// 管理员响应使用外层的展示字段。
	Usage *UpstreamUsageInfo `json:"-"`
}

// UpstreamUsageMetrics 按适配器和错误类别记录进程内的查询次数。
type UpstreamUsageMetrics struct {
	Counts map[string]int64 `json:"counts"`
}

// UpstreamUsageAdapterOption 用于前端或诊断页面展示可用适配器。
type UpstreamUsageAdapterOption struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
