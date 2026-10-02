package openaiattempt

import (
	"strings"

	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/gin-gonic/gin"
)

// ResolveOpenAIUpstreamEndpoint 返回 OpenAI 兼容提供商实际使用的上游端点。
// 同一入站可选择 Chat 或 Responses，因此优先读取转发结果，缺失时读取当前尝试上下文，再使用平台规范端点。
func ResolveOpenAIUpstreamEndpoint(c *gin.Context, provider *gatewayprovider.ExecutionProvider, result *forwardcore.OpenAIResult) string {
	if result != nil {
		if endpoint := strings.TrimSpace(result.UpstreamEndpoint); endpoint != "" {
			return endpoint
		}
	}
	if endpoint := gatewayhttp.GetActualOpenAIUpstreamEndpoint(c); endpoint != "" {
		return endpoint
	}
	return gatewayhttp.GetUpstreamEndpoint(c, provider.Record.Platform)
}
