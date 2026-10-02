package httpapi

import (
	"net/http"
	"strings"

	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"
)

// CredentialFailoverClientResponse 为已分类的凭据失败生成客户端响应。
func CredentialFailoverClientResponse(failoverErr *forwardcore.UpstreamFailoverError) (int, string) {
	if failoverErr != nil && failoverErr.Reason == forwardcore.OpenAIUpstreamAccessStateReason && strings.TrimSpace(failoverErr.ClientMessage) != "" {
		status := failoverErr.ClientStatusCode
		if status <= 0 {
			status = http.StatusServiceUnavailable
		}
		return status, failoverErr.ClientMessage
	}
	if failoverErr != nil && failoverErr.Reason == forwardcore.AntigravityCredentialRejectedReason {
		return http.StatusBadGateway, forwardcore.AntigravityCredentialRejectedClientMessage
	}
	return http.StatusServiceUnavailable, forwardcore.GrokCredentialUnavailableClientMessage
}
