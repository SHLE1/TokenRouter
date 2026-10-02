package completion_test

import (
	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/modelidentity"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"
)

// modelCatalogFixture 构造尚未启动的模型目录，供测试使用。
type modelCatalogFixture struct {
	pricingData map[string]*pricing.CatalogModelPricing
}

func newModelCatalogFixture(fixture modelCatalogFixture) *provider.Service {
	return provider.NewServiceFromSnapshot(provider.Options{
		ModelLookupCandidates: modelidentity.CandidatesFactory,
	}, nil, provider.Snapshot{Data: fixture.pricingData})
}
