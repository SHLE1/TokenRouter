package pricing

// ApplyConfigPrice 在价格副本上应用价卡覆盖，图片价格可为 nil。
func ApplyConfigPrice(pricing *ModelPricing, configPricing *ModelPricingEntry) *ModelPricing {
	if configPricing == nil {
		return pricing
	}
	// 复制目录价格后应用覆盖值。
	cloned := *pricing
	pricing = &cloned
	ApplyConfigTokenPriceOverrides(pricing, configPricing)
	multiplier, configured := NormalizedPriceMultiplier(configPricing)
	if configured {
		pricing = MultiplyModelPricing(pricing, multiplier)
	}
	ApplyConfigFastModeMultiplier(pricing, configPricing)
	ApplyConfigFlexMultiplier(pricing, configPricing)
	if configPricing.MaxReasoningEffortMultiplier != nil {
		pricing.MaxReasoningEffortMultiplier = configPricing.MaxReasoningEffortMultiplier
	}
	return pricing
}
