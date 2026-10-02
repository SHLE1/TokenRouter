package httpapi

import (
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/httpapi/dto"
)

// ProviderWithConcurrency 保留管理端的实时并发与调度展示字段。
type ProviderWithConcurrency struct {
	*dto.Provider
	CurrentConcurrency int                           `json:"current_concurrency"`
	SchedulerScore     *ProviderSchedulerScore       `json:"scheduler_score,omitempty"`
	SchedulerScores    []ProviderSchedulerGroupScore `json:"scheduler_scores,omitempty"`
	// 以下字段仅对 Anthropic OAuth/SetupToken 提供商有效，且仅在启用相应功能时返回
	CurrentWindowCost *float64 `json:"current_window_cost,omitempty"` // 当前窗口费用
	ActiveSessions    *int     `json:"active_sessions,omitempty"`     // 当前活跃会话数
	CurrentRPM        *int     `json:"current_rpm,omitempty"`         // 当前分钟 RPM 计数
}

// 提供商管理用例计算调度展示值，HTTP 将其编码为 JSON。
type (
	ProviderSchedulerScore      = providercore.ProviderSchedulerScore
	ProviderSchedulerGroupScore = providercore.ProviderSchedulerGroupScore
)
