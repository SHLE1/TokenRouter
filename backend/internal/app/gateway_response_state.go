package app

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
)

// provideOpenAIResponseState 构造由 app 管理的共享会话存储。
func provideOpenAIResponseState(cache session.GatewayCache) session.OpenAIWSStateStore {
	return session.NewOpenAIWSStateStore(cache, gatewayprovider.LogOpenAIWSModeInfo)
}
