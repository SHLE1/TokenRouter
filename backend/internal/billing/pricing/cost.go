package pricing

// CalculateCost 按计费模式计算费用，负倍率按零处理，结果记录所用模式。
func CalculateCost(resolved *ResolvedPricing, input CostInput) (*CostBreakdown, error) {
	// 保存时要求倍率大于 0，计费时将缓存或迁移数据中的负倍率按 0 处理。
	if input.RateMultiplier < 0 {
		input.RateMultiplier = 0
	}

	var breakdown *CostBreakdown
	var err error
	switch resolved.Mode {
	case BillingModePerRequest, BillingModeImage, BillingModeVideo:
		breakdown, err = CalculatePerRequestCost(resolved, input)
	default: // BillingModeToken
		breakdown, err = CalculateTokenCost(resolved, input)
	}
	if err == nil && breakdown != nil {
		breakdown.BillingMode = string(resolved.Mode)
		if breakdown.BillingMode == "" {
			breakdown.BillingMode = string(BillingModeToken)
		}
	}
	return breakdown, err
}
