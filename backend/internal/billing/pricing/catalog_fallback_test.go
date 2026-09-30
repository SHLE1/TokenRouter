package pricing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCatalogFallbackOrdering 约束基础型号优先与无基础型号时的稳定变体顺序。
func TestCatalogFallbackOrdering(t *testing.T) {
	base := &CatalogModelPricing{InputCostPerToken: 5e-6}
	first := &CatalogModelPricing{InputCostPerToken: 6e-6}
	second := &CatalogModelPricing{InputCostPerToken: 7e-6}
	for _, withBase := range []bool{false, true} {
		query := &CatalogQuery{Entries: map[string]*CatalogModelPricing{
			"claude-opus-4-6-20260101": first,
			"claude-opus-4-6-20260201": second,
		}}
		want := first
		if withBase {
			query.Entries["claude-opus-4-6"] = base
			want = base
		}
		for range 30 {
			for _, model := range []string{"claude-opus-4-6-thinking", "claude-opus-4-6-20990101"} {
				require.Same(t, want, query.GetModelPricing(model))
			}
		}
	}
}
