package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"

	"github.com/TokenFlux/TokenRouter/internal/gateway/errorpolicy"
	"github.com/gin-gonic/gin"
)

// OpenAIFailoverError 保存错误展示数据，规则匹配使用原始响应，公开 JSON 使用展示字段。
type OpenAIFailoverError struct {
	Status, ClientStatus, CredentialStatus                                                          int
	ClientMessage, CredentialMessage, TooLargeMessage, SilentMessage, CyberMessage, UpstreamMessage string
	Headers                                                                                         http.Header
	Body                                                                                            []byte
	TooLarge, ContinuationUnsupported, Credential, CapacityShed, SilentRefusal, CyberWarning        bool
}
type ErrorRuleMatcher interface {
	MatchRule(string, int, []byte) *errorpolicy.ErrorPassthroughRule
}
type FailoverErrorHooks struct {
	Upstream       func(*gin.Context, int, string)
	SkipMonitoring func(*gin.Context)
}

// WriteOpenAIFailoverExhausted 按原顺序解释已分类错误，再匹配展示规则与默认映射。
func WriteOpenAIFailoverExhausted(c *gin.Context, failure *OpenAIFailoverError, started bool, rules ErrorRuleMatcher, hooks FailoverErrorHooks, write func(*gin.Context, int, string, string, bool)) {
	if failure == nil {
		status, kind, message := MapOpenAIUpstreamError(http.StatusBadGateway, gatewayLocale(c))
		write(c, status, kind, message, started)
		return
	}
	if failure.TooLarge {
		hooks.Upstream(c, http.StatusRequestEntityTooLarge, failure.TooLargeMessage)
		write(c, http.StatusRequestEntityTooLarge, "invalid_request_error", locale.ErrorText(gatewayLocale(c), "HTTP_413", 413, failure.TooLargeMessage), started)
		return
	}
	if failure.ContinuationUnsupported {
		message := strings.TrimSpace(failure.ClientMessage)
		if message == "" {
			message = "previous_response_id requires an OpenAI API-key provider for HTTP requests"
		}
		write(c, http.StatusBadRequest, "invalid_request_error", locale.ErrorText(gatewayLocale(c), "CONTINUATION_UNSUPPORTED", 400, message), started)
		return
	}
	CopyFailoverRetryAfter(c, failure.Headers)
	if failure.Credential {
		write(c, failure.CredentialStatus, "upstream_error", locale.ErrorText(gatewayLocale(c), "UPSTREAM_AUTH_FAILED", failure.CredentialStatus, failure.CredentialMessage), started)
		return
	}
	if failure.CapacityShed && strings.TrimSpace(failure.ClientMessage) != "" {
		status := failure.ClientStatus
		if status <= 0 {
			status = http.StatusServiceUnavailable
		}
		write(c, status, "server_error", locale.ErrorText(gatewayLocale(c), "HTTP_503", status, failure.ClientMessage), started)
		return
	}
	if failure.SilentRefusal {
		hooks.Upstream(c, failure.Status, failure.SilentMessage)
		write(c, http.StatusBadGateway, "upstream_error", failure.SilentMessage, started)
		return
	}
	if failure.CyberWarning {
		hooks.Upstream(c, failure.Status, failure.CyberMessage)
		status := failure.Status
		if status < 400 || status > 599 {
			status = http.StatusBadGateway
		}
		write(c, status, "invalid_request_error", failure.CyberMessage, started)
		return
	}
	if rules != nil && len(failure.Body) > 0 {
		if rule := rules.MatchRule("openai", failure.Status, failure.Body); rule != nil {
			status := failure.Status
			if !rule.PassthroughCode && rule.ResponseCode != nil {
				status = *rule.ResponseCode
			}
			message := failure.UpstreamMessage
			if !rule.PassthroughBody && rule.CustomMessage != nil {
				message = rule.DisplayMessage(gatewayLocale(c))
			}
			if rule.SkipMonitoring {
				hooks.SkipMonitoring(c)
			}
			write(c, status, "upstream_error", message, started)
			return
		}
	}
	hooks.Upstream(c, failure.Status, failure.UpstreamMessage)
	status, kind, message := MapOpenAIUpstreamError(failure.Status, gatewayLocale(c))
	write(c, status, kind, message, started)
}

func CopyFailoverRetryAfter(c *gin.Context, headers http.Header) {
	if c == nil || headers == nil {
		return
	}
	retryAfter := strings.TrimSpace(headers.Get("Retry-After"))
	if retryAfter == "" || len(retryAfter) > 128 || strings.ContainsAny(retryAfter, "\r\n") || !IsSafeRetryAfter(retryAfter) {
		return
	}
	c.Header("Retry-After", retryAfter)
}

func IsSafeRetryAfter(value string) bool {
	digitsOnly := true
	for _, char := range value {
		if char < '0' || char > '9' {
			digitsOnly = false
			break
		}
	}
	if digitsOnly {
		seconds, err := strconv.ParseUint(value, 10, 32)
		return err == nil && seconds <= uint64((7*24*time.Hour)/time.Second)
	}
	retryAt, err := http.ParseTime(value)
	if err != nil {
		return false
	}
	return !retryAt.After(time.Now().Add(7 * 24 * time.Hour))
}

func MapOpenAIUpstreamError(statusCode int, language ...string) (int, string, string) {
	status, kind, message := mapOpenAIUpstreamError(statusCode)
	if len(language) == 0 {
		return status, kind, message
	}
	reason := "UPSTREAM_REQUEST_FAILED"
	switch statusCode {
	case 401:
		reason = "UPSTREAM_AUTH_FAILED"
	case 403:
		reason = "UPSTREAM_FORBIDDEN"
	case 429:
		reason = "UPSTREAM_RATE_LIMITED"
	case 529:
		reason = "UPSTREAM_OVERLOADED"
	case 500, 502, 503, 504:
		reason = "UPSTREAM_UNAVAILABLE"
	}
	return status, kind, locale.ErrorText(language[0], reason, status, message)
}

// mapOpenAIUpstreamError 将上游状态映射为平台生成的默认提示。
func mapOpenAIUpstreamError(statusCode int) (int, string, string) {
	switch statusCode {
	case 401:
		return http.StatusBadGateway, "upstream_error", "Upstream authentication failed, please contact administrator"
	case 403:
		return http.StatusBadGateway, "upstream_error", "Upstream access forbidden, please contact administrator"
	case 429:
		return http.StatusTooManyRequests, "rate_limit_error", "Upstream rate limit exceeded, please retry later"
	case 529:
		return http.StatusServiceUnavailable, "upstream_error", "Upstream service overloaded, please retry later"
	case 500, 502, 503, 504:
		return http.StatusBadGateway, "upstream_error", "Upstream service temporarily unavailable"
	default:
		return http.StatusBadGateway, "upstream_error", "Upstream request failed"
	}
}
