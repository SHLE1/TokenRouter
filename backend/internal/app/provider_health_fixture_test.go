package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	gatewaytestkit "github.com/TokenFlux/TokenRouter/internal/gateway/testkit"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// newAppHealthObserverFixture 组合健康状态组件和当前测试的替身。
func newAppHealthObserverFixture(store gatewayprovider.ExecutionProviderStore, cfg *config.Config) *provideradapter.UpstreamHealth {
	options := provider.HealthOptions{}
	if cfg != nil {
		options.UnauthorizedCooldownMinutes = cfg.RateLimit.OAuth401CooldownMinutes
		options.OverloadMinutes = cfg.RateLimit.OverloadCooldownMinutes
		options.CNIntervalMinutes = cfg.Gateway.CNProviders.IntervalMinutes
	}
	return gatewaytestkit.NewHealthObserver(gatewaytestkit.HealthInput{Store: store, Options: options})
}
