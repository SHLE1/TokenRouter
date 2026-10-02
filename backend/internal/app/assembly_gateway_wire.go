//go:build wireinject

package app

import (
	gatewaysession "github.com/TokenFlux/TokenRouter/internal/gateway/session"

	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"

	gatewaypg "github.com/TokenFlux/TokenRouter/internal/gateway/postgres"

	gatewayredis "github.com/TokenFlux/TokenRouter/internal/gateway/rediscache"

	"github.com/google/wire"
)

// gatewayAssemblyProviders 汇总网关协议入口和执行器的 Wire provider。
var gatewayAssemblyProviders = wire.NewSet(
	gatewaysession.NewDigestSessionStore,
	provideQoderRuntime,
	provideBackendMode,
	provideMediaHTTP,
	provideMediaRuntime,
	provideAuxiliaryHTTP,
	provideLiveHTTP,
	provideLiveExecution,
	gatewaySearchProviders,
	provideMessageHTTPBindings,
	provideMessageAttemptRuntime,
	provideMessagesHTTP,
	provideGatewayRequestActivity,
	provideModelsHTTP,
	provideResponsesWSHTTP,
	provideOpenAITextHTTP,
	provideOpenAITextAttemptRuntime,
	provideOpenAIAttemptBindings,
	provideUnifiedTextExecutor,
	provideOpenAITokensHTTP,
	provideGeminiNativeHTTP,
	provideCompatibleTextHTTP,
	provideQoderCompatibleHTTP,
	provideCountTokensHTTP,
	provideGatewayPromptPolicy,
	ProvideGatewayCompletionRecorders,
	gatewayredis.NewGatewayCache,
	gatewayErrorRulesProviders,
	provideUsageRecordWorkerPool,
	provideQoderRequestActivity,
	provideQoderChat,
	provideExecutionProviderStore,
	provideFundingAdmission,
	gatewayExecutionProviders,
	provideOpenAIHTTPResources,
	provideCyberBlocks,
	provideCyberHTTP,
	provideGatewayRouteMiddleware,
	provideGatewayAdminRules,
	provideGatewaySettings,
	provideGatewayRuntimeReaders,
	gatewayhttp.NewRuntimeSettingsHandler,
)
var gatewayErrorRulesProviders = wire.NewSet(provideGatewayErrorRules, gatewaypg.NewErrorPassthroughRepository, gatewayredis.NewErrorPassthroughCache, gatewayhttp.NewErrorPassthroughHandler)
