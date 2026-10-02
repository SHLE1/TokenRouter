package completion_test

import (
	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/billing/testkit"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"
)

// NewBillingService 根据测试配置构造计费服务。
func NewBillingService(cfg *config.Config, catalog *provider.Service) *billing.Calculator {
	return newBillingServiceWithPrices(cfg, catalog, nil)
}

func newBillingServiceWithPrices(cfg *config.Config, catalog *provider.Service, prices map[string]*pricing.ModelPricing) *billing.Calculator {
	multiplier := 0.0
	if cfg != nil {
		multiplier = cfg.Default.RateMultiplier
	}
	return testkit.Calculator(multiplier, catalog, prices)
}
