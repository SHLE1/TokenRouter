package billing_test

import (
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	billingadapter "github.com/TokenFlux/TokenRouter/internal/billing/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/media"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/modelidentity"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/stretchr/testify/require"
)

// catalogFixture 显式构造尚未启动的目录输入，不复制任何生产算法或运行状态。
type catalogFixture struct {
	pricingData map[string]*pricing.CatalogModelPricing
}

func newCatalogFixture(fixture catalogFixture) *billingadapter.PricingService {
	return billingadapter.NewPricingServiceFromSnapshot(billingadapter.Options{
		DefaultOpenAIModel:    openai.DefaultTestModel,
		IsImageModel:          media.IsImageGenerationModel,
		ModelLookupCandidates: modelidentity.CandidatesFactory,
	}, nil, billingadapter.Snapshot{Data: fixture.pricingData})
}

// newOfflineCatalogFixture 在已登记的适配夹具中加载发布目录，不让领域测试依赖具体加载器。
func newOfflineCatalogFixture(t *testing.T) *billingadapter.PricingService {
	t.Helper()
	service := billingadapter.NewPricingService(billingadapter.Options{DataDir: t.TempDir(), FallbackFile: "../../resources/model-pricing/model_pricing_supplements.json"}, nil)
	require.NoError(t, service.Initialize())
	return service
}
