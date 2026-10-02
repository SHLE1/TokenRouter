package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/gateway/admission"
)

// newFundingAdmissionFixture 复用夹具创建的资金缓存，RPM 后端保持未配置。
func newFundingAdmissionFixture(funds *billing.Eligibility, cfg *config.Config) *admission.FundingAdmission {
	return admission.NewFundingAdmission(funds, nil)
}

func billingEligibilityFixtureOptions(c *config.Config) billing.EligibilityOptions {
	return billing.EligibilityOptions{Billing: billing.BillingOptions{MinimumBalanceReserve: c.Billing.MinimumBalanceReserve, CircuitBreaker: billing.CircuitBreakerOptions{Enabled: c.Billing.CircuitBreaker.Enabled, FailureThreshold: c.Billing.CircuitBreaker.FailureThreshold, ResetTimeoutSeconds: c.Billing.CircuitBreaker.ResetTimeoutSeconds, HalfOpenRequests: c.Billing.CircuitBreaker.HalfOpenRequests}}}
}

// newBillingEligibilityFixture 构造资金缓存，绑定配置读取和异步回填函数。
func newBillingEligibilityFixture(cfg *config.Config) *billing.Eligibility {
	return billing.NewEligibility(nil, httpFixtureBalances{}, nil, func() billing.EligibilityOptions { return billingEligibilityFixtureOptions(cfg) }, nil, func(_ string, fn func()) { go fn() })
}

// httpFixtureBalances 为协议和重试测试提供可消费余额，资金准入使用生产实现。
type httpFixtureBalances struct{}

func (httpFixtureBalances) GetByID(_ context.Context, id int64) (*billing.UserSummary, error) {
	return &billing.UserSummary{ID: id, Balance: 1000}, nil
}
