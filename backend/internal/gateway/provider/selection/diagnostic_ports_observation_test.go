package selection

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	schedulercore "github.com/TokenFlux/TokenRouter/internal/scheduler"
)

func (s *diagnosticScope) providers(values []*gatewayprovider.ExecutionProvider) []*schedulercore.DiagnosticProvider {
	if values == nil {
		return nil
	}
	out := make([]*schedulercore.DiagnosticProvider, len(values))
	for i, v := range values {
		out[i] = s.provider(v)
	}
	return out
}
