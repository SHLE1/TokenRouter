package selection

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

// ResolveTransport 根据当前提供商和启动配置调用 egress 的传输选择规则。
func (s *Compatible) ResolveTransport(value *gatewayprovider.ExecutionProvider) egress.OpenAIWSProtocolDecision {
	view := gatewayprovider.ExecutionProtocolRecord(value)
	if view !=
		nil {
		view.Concurrency = value.Record.Concurrency
	}
	return gatewayprovider.ResolveOpenAIWSTransport(view)
}
