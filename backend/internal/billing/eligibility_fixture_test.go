//go:build unit

package billing

// newEligibilityForTest 直接构造唯一资金缓存，保留原独立夹具的异步回填。
func newEligibilityForTest(cache BillingCache, users BalanceReader, keys APIKeyRateLimitLoader, options *EligibilityOptions) *Eligibility {
	return NewEligibility(cache, users, keys, func() EligibilityOptions { return *options }, nil, func(_ string, fn func()) { go fn() })
}

// updateEligibilityOptionsForTest 只更新夹具的静态配置，在执行检查前完成。
func updateEligibilityOptionsForTest(service *Eligibility, update func(*EligibilityOptions)) {
	options := service.options()
	update(&options)
	service.options = func() EligibilityOptions { return options }
}
