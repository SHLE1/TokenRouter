package provider

import (
	"net/http"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream/anthropic"
)

// QuotaWindowObservation 保存供应商解析出的配额窗口，provider 执行健康状态规则。
func QuotaWindowObservation(value *anthropic.WindowLimit) *provider.QuotaWindowObservation {
	if value == nil {
		return nil
	}
	return &provider.QuotaWindowObservation{Window: value.Window, ResetAt: value.ResetAt, FiveHourReset: value.FiveHourReset, Reason: value.Reason}
}

// SessionWindowObservation 将会话响应头转换为观测字段。
func SessionWindowObservation(headers http.Header) provider.SessionWindowObservation {
	return provider.SessionWindowObservation{Status: headers.Get("anthropic-ratelimit-unified-5h-status"), Reset: headers.Get("anthropic-ratelimit-unified-5h-reset"), Passive: anthropic.PassiveUsageFields(headers)}
}
