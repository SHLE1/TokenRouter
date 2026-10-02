package billing

type APIKeyBillingContext struct {
	Mode, Source string
	Subscription *UserSubscription
	Available    bool
}

// SubscriptionRemainingForDisplay 返回展示用剩余额度，无上限时返回 -1。
func SubscriptionRemainingForDisplay(sub *UserSubscription) float64 {
	if sub == nil {
		return 0
	}
	if (sub.DailyLimitUSD == nil || *sub.DailyLimitUSD <= 0) && (sub.WeeklyLimitUSD == nil || *sub.WeeklyLimitUSD <= 0) && (sub.MonthlyLimitUSD == nil || *sub.MonthlyLimitUSD <= 0) {
		return -1
	}
	return sub.AvailableQuotaUSD()
}
