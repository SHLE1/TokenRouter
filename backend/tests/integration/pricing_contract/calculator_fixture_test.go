package pricingcontract

import (
	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/billing/testkit"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"
)

// newCalculator 将测试配置和目录传给计价器。
func newCalculator(cfg *config.Config, catalog *provider.Service) *billing.Calculator {
	return newCalculatorWithPrices(cfg, catalog, nil)
}

func newCalculatorWithPrices(cfg *config.Config, catalog *provider.Service, prices map[string]*pricing.ModelPricing) *billing.Calculator {
	multiplier := 0.0
	if cfg != nil {
		multiplier = cfg.Default.RateMultiplier
	}
	return testkit.Calculator(multiplier, catalog, prices)
}
