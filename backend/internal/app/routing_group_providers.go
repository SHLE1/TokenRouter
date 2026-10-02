package app

import (
	"context"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// routingGroupProviders 转换提供商存储记录，平台默认模型目录按需查询。
type routingGroupProviders struct {
	Store *providerpostgres.ProviderStore
}

func (r routingGroupProviders) GetByIDs(ctx context.Context, ids []int64) ([]routing.GroupProvider, error) {
	values, err := r.Store.GetByIDs(ctx, ids)
	// 批量查询无结果时返回空数组，列表查询的 nil 结果保持为 nil。
	out := make([]routing.GroupProvider, len(values))
	for i, v := range values {
		out[i] = r.project(v)
	}
	return out, err
}

func (r routingGroupProviders) ListSchedulableByGroupID(ctx context.Context, id int64) ([]routing.CatalogueProvider, error) {
	values, err := r.Store.ListSchedulableByGroupID(ctx, id)
	return gatewayprovider.CatalogueProviders(values), err
}

func (r routingGroupProviders) project(v *provider.Record) routing.GroupProvider {
	return routing.GroupProvider{ID: v.ID, Platform: v.Platform, Type: v.Type}
}
