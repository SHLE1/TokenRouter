package pricing

import (
	"fmt"
	"math"
	"strings"
)

// ConfiguredImageUnitPrice 保持显式零价与未定价之间的区别。
func ConfiguredImageUnitPrice(resolved *ResolvedPricing, size string) (float64, bool) {
	var price float64
	var found bool
	if resolved != nil && (resolved.Mode == BillingModeImage || resolved.Mode == BillingModePerRequest) {
		price, found = GetRequestTierPriceValue(resolved, strings.TrimSpace(size))
		// 已解析的正单价也可独立传入；零价仍需显式存在性，不能由缺省值推断。
		if !found && (resolved.DefaultPerRequestPrice > 0 || resolved.ConfigPricing != nil && resolved.ConfigPricing.PerRequestPrice != nil) {
			price, found = resolved.DefaultPerRequestPrice, true
		}
	}

	return price, found
}

// ValidateImageUnitPrice 保留旧固定单张价的失败语义。
func ValidateImageUnitPrice(price float64) (float64, error) {
	if math.IsNaN(price) || math.IsInf(price, 0) || price < 0 {
		return 0, fmt.Errorf("invalid image unit price: %w", ErrModelPricingUnavailable)
	}
	return price, nil
}

// ResolveImageUnitPrice 对每个尺寸先查价卡，再查完整型号目录；缺价不返回免费。
func ResolveImageUnitPrice(resolved *ResolvedPricing, catalog *CatalogModelPricing, size string) (float64, error) {
	price, found := ConfiguredImageUnitPrice(resolved, size)
	if !found {
		price, found = DefaultImagePrice(catalog, size)
	}
	if !found {
		return 0, ErrModelPricingUnavailable
	}
	return ValidateImageUnitPrice(price)
}
