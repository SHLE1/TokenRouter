package middleware

import (
	"github.com/TokenFlux/TokenRouter/internal/apikey"
	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/gin-gonic/gin"
)

// NewAPIKeyAuthMiddleware 创建 API Key 认证中间件
// NewAPIKeyAuthMiddleware 为中间件测试组合 Key 认证和订阅检查，生产装配位于 app。
func NewAPIKeyAuthMiddleware(apiKeyService *apikey.APIKeyService, subscriptionService *billing.SubscriptionService, cfg *config.Config) keyhttp.APIKeyAuthMiddleware {
	return keyhttp.APIKeyAuthMiddleware(apiKeyAuthWithSubscription(apiKeyService, subscriptionService, cfg))
}

func apiKeyAuthWithSubscription(apiKeyService *apikey.APIKeyService, subscriptionService *billing.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return newGatewayAuthorization(apiKeyService, subscriptionService, cfg, false)
}
