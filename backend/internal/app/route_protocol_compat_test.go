package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
)

type groupClientProtocolErrorFormat = gatewayhttp.GroupClientProtocolErrorFormat

const (
	groupClientProtocolErrorAnthropic = gatewayhttp.GroupClientProtocolErrorAnthropic
	groupClientProtocolErrorOpenAI    = gatewayhttp.GroupClientProtocolErrorOpenAI
	groupClientProtocolErrorGoogle    = gatewayhttp.GroupClientProtocolErrorGoogle
)

var routeProtocol = gatewayhttp.RouteProtocol

// 测试中的协议检查使用无状态适配函数，每次读取当前请求数据。
var (
	testRouteGuards                      = gatewayhttp.NewRouteGuards(legacyRouteMiddleware(nil, nil, nil, nil, nil, &config.Config{}))
	requireGroupClientProtocol           = testRouteGuards.RequireGroupClientProtocol
	extendedRouteProtocol                = gatewayhttp.ExtendedRouteProtocol
	requireGeminiGenerateContentProtocol = testRouteGuards.RequireGeminiGenerateContentProtocol
)
