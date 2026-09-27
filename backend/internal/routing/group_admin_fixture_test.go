//go:build unit

package routing_test

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/pkg/pagination"

	"github.com/TokenFlux/TokenRouter/internal/routing"
	routingprovider "github.com/TokenFlux/TokenRouter/internal/routing/provider"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
	"github.com/TokenFlux/TokenRouter/internal/scheduler/policy"
)

// newGroupAdminForTest 保留原测试的存储副本边界及默认策略，直接构造唯一分组用例。
func newGroupAdminForTest(repo routing.GroupRepository, duplicate routing.GroupDuplicateRepository, pricingConfigs routing.GroupPricingInvalidator) *routing.GroupAdmin {
	return newGroupAdminPortsForTest(repo, duplicate, pricingConfigs, nil, nil, nil, nil)
}

// newGroupAdminPortsForTest 只装配原测试需要的窄端口，不复制管理规则。
func newGroupAdminPortsForTest(repo routing.GroupRepository, duplicate routing.GroupDuplicateRepository, pricingConfigs routing.GroupPricingInvalidator, sortOrder routing.GroupSortOrderRepository, providers routing.GroupProviders, invalidator routing.GroupAdminInvalidator, weights *policy.ConfigScoreWeights, keyReaders ...routing.GroupKeyReader) *routing.GroupAdmin {
	var keys routing.GroupKeyReader
	if len(keyReaders) > 0 {
		keys = keyReaders[0]
	}
	var duplicates routing.GroupDuplicateRepository
	if duplicate != nil {
		duplicates = groupDuplicatePortFixture{duplicate}
	}
	return routing.NewGroupAdmin(groupPortFixture{repo}, duplicates, sortOrder, providers, keys, invalidator, pricingConfigs, routing.GroupAdminOptions{
		DefaultModels: routingprovider.DefaultGroupModelCandidates,
		GlobalWeights: func(ctx context.Context) (policy.ScoreWeights, error) {
			defaults := scheduler.DefaultAdminSettingsDefaults()
			if weights != nil {
				defaults.Weights = *weights
			}
			return scheduler.LoadValidationWeights(ctx, nil, defaults)
		},
		Mutate: func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
	})
}

type groupPortFixture struct{ routing.GroupRepository }

func testGroupsFixture(values []routing.Group) []routing.Group {
	if values == nil {
		return nil
	}
	out := make([]routing.Group, len(values))
	for i := range values {
		out[i] = *routing.CloneGroup(&values[i])
	}
	return out
}

func (r groupPortFixture) Create(ctx context.Context, value *routing.Group) error {
	old := routing.CloneGroup(value)
	err := r.GroupRepository.Create(ctx, old)
	*value = *routing.CloneGroup(old)
	return err
}

func (r groupPortFixture) Update(ctx context.Context, value *routing.Group) error {
	old := routing.CloneGroup(value)
	err := r.GroupRepository.Update(ctx, old)
	*value = *routing.CloneGroup(old)
	return err
}

func (r groupPortFixture) GetByID(ctx context.Context, id int64) (*routing.Group, error) {
	value, err := r.GroupRepository.GetByID(ctx, id)
	return routing.CloneGroup(value), err
}

func (r groupPortFixture) GetByIDLite(ctx context.Context, id int64) (*routing.Group, error) {
	value, err := r.GroupRepository.GetByIDLite(ctx, id)
	return routing.CloneGroup(value), err
}

func (r groupPortFixture) List(ctx context.Context, params pagination.PaginationParams) ([]routing.Group, *pagination.PaginationResult, error) {
	v, p, e := r.GroupRepository.List(ctx, params)
	return testGroupsFixture(v), p, e
}

func (r groupPortFixture) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, status, search string, isExclusive *bool) ([]routing.Group, *pagination.PaginationResult, error) {
	v, p, e := r.GroupRepository.ListWithFilters(ctx, params, platform, status, search, isExclusive)
	return testGroupsFixture(v), p, e
}

func (r groupPortFixture) ListActive(ctx context.Context) ([]routing.Group, error) {
	v, e := r.GroupRepository.ListActive(ctx)
	return testGroupsFixture(v), e
}

type groupDuplicatePortFixture struct {
	routing.GroupDuplicateRepository
}

func (p groupDuplicatePortFixture) FindByDuplicateOperationID(ctx context.Context, id string) (*routing.Group, error) {
	value, err := p.GroupDuplicateRepository.FindByDuplicateOperationID(ctx, id)
	return routing.CloneGroup(value), err
}

func (p groupDuplicatePortFixture) CreateFromSource(ctx context.Context, value *routing.Group, id int64) error {
	copy := routing.CloneGroup(value)
	err := p.GroupDuplicateRepository.CreateFromSource(ctx, copy, id)
	*value = *routing.CloneGroup(copy)
	return err
}

// imagePricingFixture 仅构造原单张价格夹具，不计算费用。
func imagePricingFixture(prices map[string]*float64) []routing.ModelPricingEntry {
	card := routing.ModelPricingEntry{Models: []string{"*"}, BillingMode: routing.BillingModeImage}
	for tier, price := range prices {
		card.Intervals = append(card.Intervals, routing.PricingInterval{TierLabel: tier, PerRequestPrice: price})
	}
	return []routing.ModelPricingEntry{card}
}
