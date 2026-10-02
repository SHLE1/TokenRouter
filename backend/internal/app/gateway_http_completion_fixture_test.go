package app

import (
	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/gateway/completion"
	gatewaytestkit "github.com/TokenFlux/TokenRouter/internal/gateway/testkit"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/usage"
)

// newHTTPCompletionFixture 为 HTTP 测试绑定完成记录依赖。
func newHTTPCompletionFixture(cfg *config.Config, logs usage.UsageLogRepository, calculator *billing.Calculator, eligibility *billing.Eligibility, activity completion.ProviderActivity, modelConfigs *routing.PricingConfigService, health completion.HealthObserver, openAI bool) *completion.Recorder {
	f := gatewaytestkit.NewRecording(logs, &gatewaytestkit.SettlementStore{}, nil, false)
	f.Dependencies.Calculator = calculator
	f.Dependencies.Health = health
	f.GroupPolicies = modelConfigs
	f.Effects.Funds.Cache = eligibility
	f.Effects.Activity = activity
	f.Options.DefaultMultiplier = 1
	if cfg != nil {
		f.Options.DefaultMultiplier = cfg.Default.RateMultiplier
	}
	return f.Core(nil, openAI)
}
