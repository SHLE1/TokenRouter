package pricing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergePriceCardsHighestIndependentBuckets(t *testing.T) {
	// 不同金额桶分别取高，输入价卡及显式零价的指针均不得被修改。
	inputA, outputA, inputB, outputB, zero := 3.0, 15.0, 5.0, 10.0, 0.0
	entries := []ModelPricingEntry{
		{ID: 1, Models: []string{" Claude-X ", "gpt-x"}, InputPrice: &inputA, OutputPrice: &outputA, CacheReadPrice: &zero},
		{ID: 2, Models: []string{"claude-x"}, InputPrice: &inputB, OutputPrice: &outputB, CacheReadPrice: &zero},
	}
	merged, conflicts := MergePriceCards(entries)
	require.Empty(t, conflicts)
	require.Len(t, merged, 2)
	require.Equal(t, []string{"claude-x"}, merged[0].Models)
	require.Equal(t, 5.0, *merged[0].InputPrice)
	require.Equal(t, 15.0, *merged[0].OutputPrice)
	require.NotNil(t, merged[0].CacheReadPrice)
	require.Zero(t, *merged[0].CacheReadPrice)
	require.Nil(t, merged[0].CacheWritePrice)
	require.Equal(t, 3.0, *entries[0].InputPrice)
	*merged[0].InputPrice = 99
	require.Equal(t, 5.0, inputB)
}

func TestMergePriceCardsRejectsIncomparableRules(t *testing.T) {
	price, other, multiplier := 1.0, 2.0, 1.5
	base := ModelPricingEntry{ID: 1, Models: []string{"model"}, InputPrice: &price}
	for _, tc := range []struct {
		name string
		edit func(*ModelPricingEntry)
	}{
		{"inheritance", func(p *ModelPricingEntry) { p.InputPrice = nil }},
		{"mode", func(p *ModelPricingEntry) { p.BillingMode = BillingModeImage; p.PerRequestPrice = &other }},
		{"multiplier", func(p *ModelPricingEntry) { p.FastMultiplier = &multiplier }},
		{"interval", func(p *ModelPricingEntry) { p.Intervals = []PricingInterval{{MinTokens: 100, InputPrice: &other}} }},
		{"time", func(p *ModelPricingEntry) { p.TimePricing = &TimePricingConfig{Timezone: "UTC"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := base.Clone()
			candidate.ID = 2
			tc.edit(&candidate)
			merged, conflicts := MergePriceCards([]ModelPricingEntry{base, candidate})
			require.Nil(t, merged)
			require.Len(t, conflicts, 1)
			require.Equal(t, []int64{1, 2}, conflicts[0].EntryIDs)
		})
	}
}

func TestMergePriceCardsIntervalsAndOverlappingPatterns(t *testing.T) {
	// 相同区间可以合并价格；不同匹配范围必须在迁移前由管理员消除歧义。
	low, high := 1.0, 2.0
	left := ModelPricingEntry{Models: []string{"model"}, Intervals: []PricingInterval{{InputPrice: &low}}}
	right := ModelPricingEntry{Models: []string{"model"}, Intervals: []PricingInterval{{InputPrice: &high}}}
	merged, conflicts := MergePriceCards([]ModelPricingEntry{left, right})
	require.Empty(t, conflicts)
	require.Equal(t, high, *merged[0].Intervals[0].InputPrice)
	right.Models = []string{"model*"}
	merged, conflicts = MergePriceCards([]ModelPricingEntry{left, right})
	require.Nil(t, merged)
	require.Equal(t, "overlapping model patterns", conflicts[0].Reason)
}
