package provider

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// AccountStatsSource 提供共享价格配置只读投影，账号成本规则由 billing 决定。
type AccountStatsSource struct{ Service *routing.PricingConfigService }

func (s AccountStatsSource) AccountStatsGroup(ctx context.Context, id int64) (*billing.AccountStatsPricingConfig, error) {
	pricingConfig, err := s.Service.GetPricingConfigForGroup(ctx, id)
	if err != nil || pricingConfig == nil {
		return nil, err
	}
	return &billing.AccountStatsPricingConfig{Rules: pricingConfig.AccountStatsPricingRules}, nil
}
