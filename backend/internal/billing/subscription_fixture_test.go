package billing_test

import (
	"context"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/pkg/timezone"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	billingpostgres "github.com/TokenFlux/TokenRouter/internal/billing/postgres"
)

// subscriptionSelectionGroupFixture 拒绝订阅选择测试中未预期的分组回源。
type subscriptionSelectionGroupFixture struct{}

func (subscriptionSelectionGroupFixture) GetByIDLite(context.Context, int64) (*billing.SubscriptionPlanGroup, error) {
	panic("unexpected GetByIDLite call")
}

// newSubscriptionServiceForTest 只绑定原生用例与原无数据库事务适配。
func newSubscriptionServiceForTest(repo billing.UserSubscriptionRepository) *billing.SubscriptionService {
	return billing.NewSubscriptionService(subscriptionSelectionGroupFixture{}, repo, billingpostgres.NewSubscriptionMutations(nil), subscriptionClockFixture())
}

// subscriptionClockFixture 返回使用 UTC 的测试时钟和日期对象。
func subscriptionClockFixture() billing.DateRuntime {
	calendar := timezone.NewCalendar(time.UTC)
	return billing.DateRuntime{Now: func() time.Time { return time.Now().UTC() }, Calendar: &calendar}
}
