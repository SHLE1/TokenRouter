package app

import (
	"github.com/TokenFlux/TokenRouter/internal/gateway/completion"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
	wshttp "github.com/TokenFlux/TokenRouter/internal/gateway/ws/httpapi"
)

// gatewayExecutionFixture 保存测试构造的执行组件和输入。
type gatewayExecutionFixture struct {
	Text       *gatewayhttp.OpenAITextExecutor
	Requests   *gatewayhttp.OpenAIRequests
	Responses  *gatewayhttp.OpenAIResponsesExecutor
	WebSockets *wshttp.OpenAIWebSocketExecutor
	Grok       *gatewayhttp.GrokExecutor
	Auxiliary  *gatewayhttp.OpenAIAuxiliary
	Recorder   *completion.Recorder
	Blocks     *session.CyberBlocks
	Cache      session.GatewayCache
	Planner    *gatewayadapter.RoutePlanner
	Background func(string, func()) bool
}
