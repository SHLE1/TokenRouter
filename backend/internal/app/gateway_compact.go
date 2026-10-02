package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

// provideCompactExecutor 为上下文恢复执行器提供所需的静态参数。
func provideCompactExecutor(cfg *config.Config) *gatewayhttp.CompactExecutor {
	value := &gatewayhttp.CompactExecutor{}
	if cfg != nil {
		value.Models = gatewayadapter.CompactModels{Default: cfg.Gateway.OpenAICompactModel}
		value.LogBody = cfg.Gateway.LogUpstreamErrorBody
		value.LogBodyMaxBytes = cfg.Gateway.LogUpstreamErrorBodyMaxBytes
	}
	return value
}
