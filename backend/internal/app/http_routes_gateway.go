package app

import (
	batchhttp "github.com/TokenFlux/TokenRouter/internal/batchimage/httpapi"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	wshttp "github.com/TokenFlux/TokenRouter/internal/gateway/ws/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/usage/httpapi"
	"github.com/gin-gonic/gin"
)

// provideGatewayRouteMount 为路由注册函数绑定已构造的网关端点。
func provideGatewayRouteMount(eCountTokensHTTP *gatewayhttp.CountTokensHandler,
	eQoderCompatibleHTTP *gatewayhttp.QoderCompatibleHandler,
	eCompatibleTextHTTP *gatewayhttp.CompatibleTextHandler,
	eGeminiNativeHTTP *gatewayhttp.GeminiNativeHandler,
	eOpenAITextHTTP *gatewayhttp.OpenAITextHandler,
	eOpenAITokensHTTP *gatewayhttp.OpenAITokensHandler,
	eResponsesWSHTTP *wshttp.ResponsesWSHandler,
	eModelsHTTP *gatewayhttp.ModelsHandler,
	eMessagesHTTP *gatewayhttp.MessagesHandler,
	eMediaHTTP *gatewayhttp.MediaHandler,
	eAuxiliaryHTTP *gatewayhttp.AuxiliaryHandler,
	eLiveHTTP *gatewayhttp.LiveHandler,
	eSearchHTTP *gatewayhttp.SearchHandler,
	ePublicUsage *httpapi.PublicUsageHandler,
	eQoderChat *gatewayhttp.QoderChatHandler,
	batch *batchhttp.BatchImageHandler,
	options gatewayhttp.RouteMiddleware,
) gatewayRouteMount {
	return func(r *gin.Engine) {
		gatewayhttp.RegisterGatewayRoutes(r, gatewayhttp.RouteEndpoints{CountTokens: eCountTokensHTTP, QoderCompatible: eQoderCompatibleHTTP, CompatibleText: eCompatibleTextHTTP, GeminiNative: eGeminiNativeHTTP, OpenAIText: eOpenAITextHTTP, OpenAITokens: eOpenAITokensHTTP, ResponsesWS: eResponsesWSHTTP.ResponsesWebSocket, Models: eModelsHTTP, Messages: eMessagesHTTP, Media: eMediaHTTP, Auxiliary: eAuxiliaryHTTP, Live: eLiveHTTP, Search: eSearchHTTP, PublicUsage: ePublicUsage.Usage, QoderChat: eQoderChat.ChatCompletions}, options, func(group *gin.RouterGroup) { batchhttp.RegisterGatewayRoutes(group, batch) })
	}
}
