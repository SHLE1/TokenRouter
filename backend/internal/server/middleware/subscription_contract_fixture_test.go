package middleware

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	billingpostgres "github.com/TokenFlux/TokenRouter/internal/billing/postgres"
)

// subscriptionAuthGroups 为未配置分组来源的测试返回空查询结果。
type subscriptionAuthGroups struct{}

func (subscriptionAuthGroups) GetByIDLite(context.Context, int64) (*billing.SubscriptionPlanGroup, error) {
	return nil, nil
}

// newSubscriptionAuthFixture 为认证测试构造订阅服务。
func newSubscriptionAuthFixture(repo billing.UserSubscriptionRepository) *billing.SubscriptionService {
	return billing.NewSubscriptionService(subscriptionAuthGroups{}, repo, billingpostgres.NewSubscriptionMutations(nil))
}
