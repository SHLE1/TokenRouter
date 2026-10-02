package app

import (
	"time"

	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/selection"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
)

// provideCodexTurnStateHeaders 为 HTTP 响应绑定共享的来源表和粘性会话 TTL。
func provideCodexTurnStateHeaders(choices *selection.Compatible) *gatewayhttp.CodexTurnStateHeaders {
	return &gatewayhttp.CodexTurnStateHeaders{Origins: session.NewCodexTurnOrigins(time.Now), TTL: choices.SessionStickyTTL}
}
