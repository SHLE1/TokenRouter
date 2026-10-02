package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/apikey"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
)

// GetExecutionAPIKey 从请求上下文读取已认证的 API Key。
func GetExecutionAPIKey(c interface{ Get(string) (any, bool) }) *apikey.APIKey {
	if c == nil {
		return nil
	}
	v, exists := c.Get("api_key")
	if !exists {
		return nil
	}
	apiKey, _ := v.(*apikey.APIKey)
	return apiKey
}

func OpenAIClientPolicyForbiddenMessage(result providercore.CodexClientRestrictionDetectionResult) string {
	// 按客户端策略返回拒绝原因。
	if result.Policy == providercore.OpenAIOAuthClientPolicyCodexOnly {
		return "This provider only allows Codex official clients"
	}
	if result.Policy == providercore.OpenAIOAuthClientPolicyTLSRouterMatchedOnly {
		return "This provider only allows clients matched by the configured TLS router"
	}
	return "This provider only allows configured OpenAI OAuth clients"
}
