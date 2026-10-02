package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/site"
)

// siteBillingSubscriptions 将 billing 的有效订阅转换为公告资格判断需要的数据。
type siteBillingSubscriptions struct {
	repo billing.UserSubscriptionRepository
}

func (a siteBillingSubscriptions) ListActiveByUserID(ctx context.Context, id int64) ([]site.SubscriptionSnapshot, error) {
	subs, err := a.repo.ListActiveByUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]site.SubscriptionSnapshot, len(subs))
	for i, s := range subs {
		out[i] = site.SubscriptionSnapshot{PlanID: s.PlanID}
	}
	return out, nil
}

func provideAnnouncementSubscriptions(repo billing.UserSubscriptionRepository) site.SubscriptionReader {
	return siteBillingSubscriptions{repo: repo}
}
