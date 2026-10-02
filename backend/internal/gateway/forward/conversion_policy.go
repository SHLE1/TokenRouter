package forward

import (
	"github.com/TokenFlux/TokenRouter/internal/protocol/bridge"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// ConversionOptionsForModel 根据模型策略生成协议转换选项。
func ConversionOptionsForModel(model string) bridge.RequestOptions {
	return bridge.RequestOptions{
		DropSampling:      capability.ResponsesBridgeDropsSampling(model),
		SupportsMaxEffort: capability.ResponsesBridgeSupportsMaxEffort(model),
	}
}
