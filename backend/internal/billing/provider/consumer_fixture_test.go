package provider

import (
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/gateway/media"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/modelidentity"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/stretchr/testify/require"
)

// 测试准备发生在启动前，直接使用所属模块状态，不复制旧服务或增加生产接口。
type pricingServiceFixture struct {
	options     Options
	pricingData map[string]*pricing.CatalogModelPricing
}

func newPricingServiceFixture(fixture pricingServiceFixture) *PricingService {
	options := fixture.options
	options.DefaultOpenAIModel = openai.DefaultTestModel
	options.IsImageModel = media.IsImageGenerationModel
	options.ModelLookupCandidates = modelidentity.CandidatesFactory
	return NewPricingServiceFromSnapshot(options, nil, Snapshot{Data: fixture.pricingData})
}

func setPricingFixtureData(service *PricingService, data map[string]*pricing.CatalogModelPricing) {
	service.pricingData = data
}

func mutatePricingFixture(service *PricingService, change func(map[string]*pricing.CatalogModelPricing)) {
	data := service.Snapshot().Data
	change(data)
	setPricingFixtureData(service, data)
}

func setPricingFixtureRemote(service *PricingService, remote PricingRemoteClient) {
	service.remoteClient = remote
}

func newStubPricingServiceFromJSON(t *testing.T, body string) *PricingService {
	t.Helper()
	service := newPricingServiceFixture(pricingServiceFixture{})
	data, err := parsePricingFixture([]byte(body))
	require.NoError(t, err)
	setPricingFixtureData(service, data)
	return service
}

// newBillingFixture 目录和计费测试组合生产计算器，并注入缺省倍率和时钟。
func newBillingFixture(catalog *PricingService) *billing.Calculator {
	var source billing.PriceCatalog
	if catalog != nil {
		source = catalog
	}
	warnings := &PricingWarnings{}
	return billing.NewCalculator(source, billing.CalculatorOptions{
		ModelPolicy:     modelidentity.PricingPolicy,
		Now:             time.Now,
		LoadLocation:    LoadPricingLocation,
		FallbackWarning: warnings.Fallback,
	})
}

// parsePricingFixture 纯价格测试使用生产解析器，不启动目录服务。
func parsePricingFixture(body []byte) (map[string]*pricing.CatalogModelPricing, error) {
	raw, err := pricing.DecodeCatalogEntries(body)
	if err != nil {
		return nil, err
	}
	values, diagnostics, err := pricing.ParsePricingEntries(raw)
	if validationErr := diagnostics.ValidationError(); validationErr != nil {
		return nil, validationErr
	}
	return values, err
}

// newOfflinePricingFixture 使用真实内嵌目录和发布补充文件验证默认价格。
func newOfflinePricingFixture(t *testing.T) *PricingService {
	t.Helper()
	service := newPricingServiceFixture(pricingServiceFixture{options: Options{DataDir: t.TempDir(), FallbackFile: "../../../resources/model-pricing/model_pricing_supplements.json"}})
	require.NoError(t, service.Initialize())
	return service
}
