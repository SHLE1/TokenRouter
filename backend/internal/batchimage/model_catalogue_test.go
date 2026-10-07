package batchimage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/batchimage"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/stretchr/testify/require"
)

// batchCataloguePolicyReads 统计批量模型列表读取策略和计费来源的次数。
type batchCataloguePolicyReads struct {
	*routing.PricingConfigService
	policyReads, pricingReads int
}

// GetGroupPolicy 统计策略读取次数。
func (p *batchCataloguePolicyReads) GetGroupPolicy(ctx context.Context, id int64) (*routing.GroupPolicyView, error) {
	p.policyReads++
	return p.PricingConfigService.GetGroupPolicy(ctx, id)
}

// GetPricingConfigForGroup 统计计费来源读取次数。
func (p *batchCataloguePolicyReads) GetPricingConfigForGroup(ctx context.Context, id int64) (*routing.PricingConfig, error) {
	p.pricingReads++
	return p.PricingConfigService.GetPricingConfigForGroup(ctx, id)
}

// TestBatchImageCatalogueResolvesBeforeModalities 覆盖通配来源、分组改写和最终输出模态。
func TestBatchImageCatalogueResolvesBeforeModalities(t *testing.T) {
	for _, tc := range []struct {
		name            string
		groupMapping    map[string]string
		providerMapping map[string]any
		whitelist       []string
		final           string
		visible         bool
	}{
		{name: "provider wildcard", providerMapping: map[string]any{"gpt-*": "catalog-image"}, whitelist: []string{"catalog-image"}, final: "catalog-image", visible: true},
		{name: "group wildcard", groupMapping: map[string]string{"gpt-*": "catalog-image"}, whitelist: []string{"catalog-image"}, final: "catalog-image", visible: true},
		{name: "two mapping stages", groupMapping: map[string]string{"gpt-*": "route-model"}, providerMapping: map[string]any{"route-*": "catalog-image"}, whitelist: []string{"catalog-image"}, final: "catalog-image", visible: true},
		{name: "unknown configured final", providerMapping: map[string]any{"gpt-*": "nano-banana-pro"}, whitelist: []string{"nano-banana-pro"}, final: "nano-banana-pro", visible: true},
		{name: "mapped text final", providerMapping: map[string]any{"gpt-*": "catalog-text"}, whitelist: []string{"catalog-text"}, final: "catalog-text"},
		{name: "provider restriction", providerMapping: map[string]any{"gpt-*": "catalog-image"}, whitelist: []string{"catalog-text"}, final: "catalog-image"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _, _, _, _ := newTestBatchImagePublicService(true)
			owner := testBatchImageOwner()
			group := *owner.GroupID
			value := testBatchImageMappedProvider(303, "apikey", tc.providerMapping)
			value.Credentials["model_whitelist"] = tc.whitelist
			svc.ProviderRepo = rebindBatchFixtureProviders(svc, &publicBatchImageProviderRepo{providers: []provider.Record{value}})
			svc.ModelIDs = func() []string { return []string{"gpt-5.4", "catalog-image", "catalog-text", "unconfigured-unknown"} }
			image, text := []string{"image"}, []string{"text"}
			svc.ModelOutputModalities = func(id string) *[]string {
				switch id {
				case "catalog-image":
					return &image
				case "gpt-5.4", "catalog-text":
					return &text
				default:
					return nil
				}
			}
			policy := newPublicPricingConfigFixture(makePublicPricingConfigFixture(routing.PricingConfig{ID: 1, Status: "active", GroupIDs: []int64{group}, BillingModelSource: routing.BillingModelSourceUpstream}, nil, routing.GroupRoutingPolicy{Enabled: true, ModelMapping: tc.groupMapping}))
			svc.PricingConfigService = policy
			price := &fakeBatchImagePricingResolver{unitPrice: 0.1}
			svc.Pricing = price
			result, err := svc.ListModels(context.Background(), owner)
			require.NoError(t, err)
			ids := []string{}
			for _, model := range result.Data {
				ids = append(ids, model.ID)
			}
			if tc.visible {
				require.Contains(t, ids, "gpt-5.4")
				require.Contains(t, price.models, tc.final)
			} else {
				require.NotContains(t, ids, "gpt-5.4")
			}
			require.NotContains(t, ids, "unconfigured-unknown")
			require.NotContains(t, ids, "catalog-text")
		})
	}
}

// TestBatchImageCatalogueUsesOnePolicySnapshot 大目录在映射后判断模态和分组限制，策略只读一次。
func TestBatchImageCatalogueUsesOnePolicySnapshot(t *testing.T) {
	for _, stage := range []string{routing.BillingModelSourceRequested, routing.BillingModelSourceGroupMapped, routing.BillingModelSourceUpstream} {
		t.Run(stage, func(t *testing.T) {
			svc, _, _, _, _, _ := newTestBatchImagePublicService(true)
			owner := testBatchImageOwner()
			group := *owner.GroupID
			allowed := map[string]string{routing.BillingModelSourceRequested: "gpt-5.4", routing.BillingModelSourceGroupMapped: "route-model", routing.BillingModelSourceUpstream: "catalog-image"}[stage]
			policies := &batchCataloguePolicyReads{PricingConfigService: newPublicPricingConfigFixture(makePublicPricingConfigFixture(routing.PricingConfig{ID: 1, Status: "active", GroupIDs: []int64{group}, BillingModelSource: routing.BillingModelSourceUpstream}, nil, routing.GroupRoutingPolicy{Enabled: true, ModelMapping: map[string]string{"gpt-*": "route-model"}, RestrictModels: true, RestrictionModelSource: stage, AllowedModels: []string{allowed}}))}
			svc.PricingConfigService = policies
			value := testBatchImageMappedProvider(303, "apikey", map[string]any{"route-*": "catalog-image"})
			value.Credentials["model_whitelist"] = []string{"catalog-image"}
			svc.ProviderRepo = rebindBatchFixtureProviders(svc, &publicBatchImageProviderRepo{providers: []provider.Record{value}})
			svc.ModelIDs = func() []string {
				ids := []string{"gpt-5.4", "catalog-image"}
				for i := 0; i < 12000; i++ {
					ids = append(ids, fmt.Sprintf("text-%d", i))
				}
				return ids
			}
			image, text := []string{"image"}, []string{"text"}
			svc.ModelOutputModalities = func(id string) *[]string {
				if id == "catalog-image" {
					return &image
				}
				return &text
			}
			svc.Pricing = &fakeBatchImagePricingResolver{unitPrice: 0.1}
			result, err := svc.ListModels(context.Background(), owner)
			require.NoError(t, err)
			ids := []string{}
			for _, model := range result.Data {
				ids = append(ids, model.ID)
			}
			require.Contains(t, ids, "gpt-5.4")
			require.NotContains(t, ids, "text-0")
			require.Equal(t, 1, policies.policyReads)
			require.Equal(t, 1, policies.pricingReads)
		})
	}
}

// TestBatchImageCatalogueRejectsMissingPrice 模态合格的映射仍需有效图片报价。
func TestBatchImageCatalogueRejectsMissingPrice(t *testing.T) {
	svc, _, _, _, _, _ := newTestBatchImagePublicService(true)
	value := testBatchImageMappedProvider(303, "apikey", map[string]any{"gpt-*": "catalog-image"})
	svc.ProviderRepo = rebindBatchFixtureProviders(svc, &publicBatchImageProviderRepo{providers: []provider.Record{value}})
	svc.ModelIDs = func() []string { return []string{"gpt-5.4"} }
	image := []string{"image"}
	svc.ModelOutputModalities = func(string) *[]string { return &image }
	svc.Pricing = &fakeBatchImagePricingResolver{err: batchimage.ErrBatchImageSettlementPricingMissing}
	result, err := svc.ListModels(context.Background(), testBatchImageOwner())
	require.NoError(t, err)
	require.Empty(t, result.Data)
}

// TestBatchImageAutomaticCandidates 自动候选按最终模态筛选，未配置的未知型号保持隐藏。
func TestBatchImageAutomaticCandidates(t *testing.T) {
	svc, _, _, _, _, _ := newTestBatchImagePublicService(true)
	value := testBatchImageMappedProvider(303, "apikey", map[string]any{"source-image": "final-text"})
	svc.ProviderRepo = rebindBatchFixtureProviders(svc, &publicBatchImageProviderRepo{providers: []provider.Record{value}})
	svc.ModelIDs = func() []string { return []string{"catalog-image", "source-image", "final-text", "unknown"} }
	image, text := []string{"image"}, []string{"text"}
	svc.ModelOutputModalities = func(id string) *[]string {
		switch id {
		case "catalog-image", "source-image":
			return &image
		case "final-text":
			return &text
		default:
			return nil
		}
	}
	svc.Pricing = &fakeBatchImagePricingResolver{unitPrice: 0.1}
	result, err := svc.ListModels(context.Background(), testBatchImageOwner())
	require.NoError(t, err)
	require.NotEmpty(t, result.Data)
	for _, model := range result.Data {
		require.Equal(t, "catalog-image", model.ID)
	}
}
