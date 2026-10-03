package dto

import (
	"encoding/json"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/stretchr/testify/require"
)

// TestMarketplacePricingJSONIncludesTokenModifiers 检查模型广场能拿到分时规则和 Max 推理倍率。
func TestMarketplacePricingJSONIncludesTokenModifiers(t *testing.T) {
	maxMultiplier := 1.5
	row := modelMarketplacePricingFromRouting(pricing.ModelDisplayPricing{
		PricingMode:                  "token",
		PriceStatus:                  "priced",
		InputPricePerToken:           1e-6,
		MaxReasoningEffortMultiplier: &maxMultiplier,
		TimePricing: &pricing.TimePricingConfig{
			Timezone:     "Asia/Shanghai",
			WeekdaysOnly: true,
			Periods:      []pricing.TimePricingPeriod{{StartTime: "20:00", EndTime: "00:00", Multiplier: 0.5}},
		},
	})
	body, err := json.Marshal(row)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"pricing_mode": "token",
		"price_status": "priced",
		"input_price_per_token": 0.000001,
		"max_reasoning_effort_multiplier": 1.5,
		"time_pricing": {
			"timezone": "Asia/Shanghai",
			"weekdays_only": true,
			"periods": [{"start_time": "20:00", "end_time": "00:00", "multiplier": 0.5}]
		}
	}`, string(body))
}

// TestMarketplacePricingJSONOmitsAbsentTokenModifiers 检查没有倍率的模型不输出这两个字段。
func TestMarketplacePricingJSONOmitsAbsentTokenModifiers(t *testing.T) {
	row := modelMarketplacePricingFromRouting(pricing.ModelDisplayPricing{PricingMode: "token", PriceStatus: "priced", InputPricePerToken: 1e-6})
	body, err := json.Marshal(row)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(body, &fields))
	require.NotContains(t, fields, "max_reasoning_effort_multiplier")
	require.NotContains(t, fields, "time_pricing")
}
