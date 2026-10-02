package app

import (
	"context"
	"log/slog"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// catalogueReader 查询指定分组或全部可调度提供商，由网关适配器转换模型数据。
func catalogueReader(store *providerpostgres.ProviderStore) func(context.Context, *int64) ([]routing.CatalogueProvider, error) {
	return func(ctx context.Context, id *int64) ([]routing.CatalogueProvider, error) {
		var values []provider.Record
		var err error
		if id != nil {
			values, err = store.ListSchedulableByGroupID(ctx, *id)
		} else {
			values, err = store.ListSchedulable(ctx)
		}
		if err != nil {
			return nil, err
		}
		return gatewayprovider.CatalogueProviders(values), nil
	}
}

// provideRequestableCatalogue 与市场、模型列表共用缓存和提供商存储。
func provideRequestableCatalogue(models *routing.ModelList, store *providerpostgres.ProviderStore, modelConfigs *routing.PricingConfigService) *routing.RequestableCatalogue {
	return &routing.RequestableCatalogue{Models: models, Read: catalogueReader(store), Resolver: routing.RequestableResolver{GroupPolicies: modelConfigs, Defaults: gatewayprovider.CatalogueDefaults(), Warn: slog.Warn}, Warn: slog.Warn}
}
