package routing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

// marketplaceLoadingModels 为部分分组返回空目录，覆盖属性和观测查询的分组筛选。
type marketplaceLoadingModels struct{}

// marketplaceLoadingProtocols 是解析器给 alias 返回的客户端协议。
var marketplaceLoadingProtocols = []capability.ProtocolID{capability.ProtocolAnthropicMessages, capability.ProtocolOpenAIChatCompletions}

func (marketplaceLoadingModels) Prefetch(context.Context) ([]routing.CatalogueProvider, bool, error) {
	return nil, false, nil
}

func (marketplaceLoadingModels) ResolveRequestableModels(_ context.Context, id *int64, _ string) routing.RequestableModelsResult {
	if *id > 2 {
		return routing.RequestableModelsResult{}
	}
	return routing.RequestableModelsResult{Models: []routing.RequestableModel{{ID: "alias", PricingModel: "priced", UpstreamModels: []string{"upstream"}, Protocols: marketplaceLoadingProtocols, NativeProtocols: marketplaceLoadingProtocols[:1]}}}
}

// marketplaceLoadingObservations 记录容量和可用率查询的分组。
type marketplaceLoadingObservations struct {
	capacityCalls      int
	capacityGroups     []int64
	availabilityGroups []int64
}

func (o *marketplaceLoadingObservations) GetGroupCapacityByIDs(_ context.Context, ids []int64) (map[int64]routing.GroupCapacitySummary, error) {
	o.capacityCalls++
	o.capacityGroups = ids
	return map[int64]routing.GroupCapacitySummary{1: {}}, nil
}

func (o *marketplaceLoadingObservations) GetSummaryByGroupIDs(_ context.Context, ids []int64, _ int, _ int, _ string, _ time.Time) (map[int64]*routing.GroupAvailabilitySummary, error) {
	o.availabilityGroups = ids
	return map[int64]*routing.GroupAvailabilitySummary{1: {WindowDays: 3}}, nil
}

// marketplaceLoadingPrices 记录报价和兼容模态查询。
type marketplaceLoadingPrices struct {
	requests      []routing.MarketplaceQuoteRequest
	modalityCalls int
}

func (p *marketplaceLoadingPrices) Quote(_ context.Context, request routing.MarketplaceQuoteRequest) pricing.ModelDisplayPricing {
	p.requests = append(p.requests, request)
	return pricing.ModelDisplayPricing{PriceStatus: "priced", InputPricePerToken: request.RateMultiplier}
}

func (p *marketplaceLoadingPrices) GetModelModalities(string) ([]string, []string) {
	p.modalityCalls++
	return []string{"text"}, []string{"text"}
}

// TestMarketplaceBatchAttributesAndOptionalCapacity 覆盖批量读取、属性失败降级、容量开关和模型协议的透传。
func TestMarketplaceBatchAttributesAndOptionalCapacity(t *testing.T) {
	for _, attributesFail := range []bool{false, true} {
		for _, includeCapacity := range []bool{false, true} {
			groups := &marketplaceGroupRepoStub{groups: []routing.Group{
				{ID: 2, ActiveProviderCount: 1, RateMultiplier: 2},
				{ID: 1, ActiveProviderCount: 1, RateMultiplier: 1, AvailabilityProbeConfig: routing.GroupAvailabilityProbeConfig{Enabled: true}},
				{ID: 3, ActiveProviderCount: 1},
				{ID: 4, ActiveProviderCount: 1, IsExclusive: true},
				{ID: 5},
			}}
			observations := &marketplaceLoadingObservations{}
			prices := &marketplaceLoadingPrices{}
			calls := 0
			name := "display"
			modalities := []string{"image"}
			options := routing.MarketplaceOptions{
				Now:          time.Now,
				Warn:         func(string, ...any) {},
				DisplayNames: func(string) map[string]string { return nil },
				Attributes: func(_ context.Context, models map[int64][]routing.RequestableModel) (map[int64]map[string]routing.EffectiveModelAttributes, error) {
					calls++
					require.Len(t, models, 2)
					for _, id := range []int64{1, 2} {
						require.Equal(t, []routing.RequestableModel{{ID: "alias", UpstreamModels: []string{"upstream"}}}, models[id])
					}
					if attributesFail {
						return nil, errors.New("database unavailable")
					}
					return map[int64]map[string]routing.EffectiveModelAttributes{1: {"alias": {Attributes: modelcatalog.Attributes{DisplayName: &name, InputModalities: &modalities}}}}, nil
				},
			}
			svc := routing.NewMarketplace(groups, nil, marketplaceLoadingModels{}, routing.RequestableResolver{}, prices, observations, observations, options)
			result, err := svc.ListPublic(context.Background(), routing.MarketplaceListOptions{IncludeCapacity: includeCapacity})
			require.NoError(t, err)
			require.Len(t, result, 2)
			require.Equal(t, int64(2), result[0].ID)
			require.Equal(t, int64(1), result[1].ID)
			require.Equal(t, 1, calls)
			require.Zero(t, prices.modalityCalls)
			require.Len(t, prices.requests, 2)
			require.Equal(t, "priced", prices.requests[0].Model)
			require.Equal(t, float64(2), result[0].Models[0].Pricing.InputPricePerToken)
			require.Equal(t, marketplaceLoadingProtocols, result[0].Models[0].Protocols)
			require.Equal(t, marketplaceLoadingProtocols[:1], result[0].Models[0].NativeProtocols)
			require.Equal(t, []int64{1}, observations.availabilityGroups)
			require.NotNil(t, result[1].Availability)
			if includeCapacity {
				require.Equal(t, 1, observations.capacityCalls)
				require.Equal(t, []int64{2, 1}, observations.capacityGroups)
				require.NotNil(t, result[1].Capacity)
			} else {
				require.Zero(t, observations.capacityCalls)
				require.Nil(t, result[1].Capacity)
			}
			if attributesFail {
				require.Nil(t, result[1].Models[0].Attributes)
				require.Nil(t, result[1].Models[0].InputModalities)
			} else {
				require.Equal(t, name, result[1].Models[0].DisplayName)
				require.Equal(t, modalities, result[1].Models[0].InputModalities)
			}
		}
	}
}
