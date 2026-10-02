package selection

import (
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// providers 将测试候选转换为诊断列表，输入为 nil 时返回 nil。
func (s *diagnosticScope) providers(values []*provider.ExecutionProvider) []*scheduler.DiagnosticProvider {
	if values == nil {
		return nil
	}
	out := make([]*scheduler.DiagnosticProvider, len(values))
	for i, v := range values {
		out[i] = s.provider(v)
	}
	return out
}
