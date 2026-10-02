package completion_test

import "github.com/TokenFlux/TokenRouter/internal/routing"

// testImageModelPricing 通过模型价卡提供测试所需的媒体价格。
func testImageModelPricing(prices map[string]*float64) []routing.ModelPricingEntry {
	return testMediaModelPricing(routing.BillingModeImage, prices)
}

func testVideoModelPricing(prices map[string]*float64) []routing.ModelPricingEntry {
	return testMediaModelPricing(routing.BillingModeVideo, prices)
}

func testMediaModelPricing(mode routing.BillingMode, prices map[string]*float64) []routing.ModelPricingEntry {
	card := routing.ModelPricingEntry{Models: []string{"*"}, BillingMode: mode}
	for tier, price := range prices {
		card.Intervals = append(card.Intervals, routing.PricingInterval{TierLabel: tier, PerRequestPrice: price})
	}
	return []routing.ModelPricingEntry{card}
}
