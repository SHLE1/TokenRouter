package provider

import (
	"net/http"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/upstream/grok"
)

// GrokRetryableOnSameProvider 标记共享 failover 循环可在同提供商重试的瞬态错误。
// 模型容量压力允许有限重试；免费额度和计费耗尽属于提供商状态，应立即切换提供商。
func GrokRetryableOnSameProvider(provider *ExecutionProvider, statusCode int, responseBody []byte) bool {
	if provider == nil || !provider.View().IsGrok() {
		return false
	}
	// 配置的错误码策略优先于池模式默认状态列表，命中后交给外层故障转移处理。
	if provider.View().IsCustomErrorCodesEnabled() && provider.View().ShouldHandleErrorCode(statusCode) {
		return false
	}
	decision := grok.ClassifyGrokUpstreamFailure(statusCode, responseBody, "")
	switch decision.Class {
	case grok.GrokFailureFreeUsage, grok.GrokFailureBilling, grok.GrokFailureCompatibility:
		// 额度和权益耗尽不能靠同提供商重放恢复。
		return false
	case grok.GrokFailureModelCapacity:
		if statusCode == http.StatusTooManyRequests {
			return true
		}
	}
	return provider.View().IsPoolMode() && provider.View().IsPoolModeRetryableStatus(statusCode)
}

func GrokSameProviderRetryMetadata(provider *ExecutionProvider, statusCode int, responseBody []byte) (bool, time.Duration, time.Time, int) {
	if !GrokRetryableOnSameProvider(provider, statusCode, responseBody) {
		return false, 0, time.Time{}, 0
	}
	decision := grok.ClassifyGrokUpstreamFailure(statusCode, responseBody, "")
	if decision.Class != grok.GrokFailureModelCapacity {
		return true, 0, time.Time{}, 0
	}
	// 每次上游尝试都会重新构造错误，因此错误上的截止时间不能覆盖整个请求；
	// 容量错误最多重放一次，首次尝试超过 30 秒时也适用。
	return true, 500 * time.Millisecond, time.Now().Add(30 * time.Second), 1
}
