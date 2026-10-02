package selection

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// modelRejectionSources 将候选提供商转换为 routing 的模型拒绝判断输入。
func modelRejectionSources(providers []gatewayprovider.ExecutionProvider) []routing.ModelRejectionSource {
	sources := make([]routing.ModelRejectionSource, len(providers))
	for i := range providers {
		value := &providers[i]
		sources[i] = gatewayprovider.ModelRejectionProvider(gatewayprovider.ExecutionRecord(value))
	}
	return sources
}
