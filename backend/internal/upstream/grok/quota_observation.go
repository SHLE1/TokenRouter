package grok

import (
	"net/http"
	"time"
)

// ParseQuotaObservation 解析额度响应头，缺少额度头的 429 响应记录状态码和观测时间。
func ParseQuotaObservation(headers http.Header, status int, now time.Time) *QuotaSnapshot {
	snapshot := ParseQuotaHeaders(headers, status)
	if snapshot == nil && status == http.StatusTooManyRequests {
		return &QuotaSnapshot{StatusCode: status, UpdatedAt: now.UTC().Format(time.RFC3339)}
	}
	return snapshot
}
