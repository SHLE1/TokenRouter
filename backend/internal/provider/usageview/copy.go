package usageview

import "github.com/TokenFlux/TokenRouter/internal/upstream/usageview"

func CloneQuotaWindow(value *QuotaWindow) *QuotaWindow { return usageview.CloneQuotaWindow(value) }
func CloneBillingSummary(value *BillingSummary) *BillingSummary {
	return usageview.CloneBillingSummary(value)
}
