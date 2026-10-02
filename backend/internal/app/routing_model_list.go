package app

import (
	"time"

	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"

	"github.com/TokenFlux/TokenRouter/internal/config"

	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// provideRoutingModelList 为模型列表缓存配置 TTL，默认十五秒。
func provideRoutingModelList(repo *providerpostgres.ProviderStore, cfg *config.Config) *routing.ModelList {
	return routing.NewModelList(catalogueReader(repo), resolveModelsListCacheTTL(cfg))
}

// resolveModelsListCacheTTL 返回配置的正数 TTL，其他情况使用十五秒。
func resolveModelsListCacheTTL(cfg *config.Config) time.Duration {
	if cfg == nil || cfg.Gateway.ModelsListCacheTTLSeconds <= 0 {
		return 15 * time.Second
	}
	return time.Duration(cfg.Gateway.ModelsListCacheTTLSeconds) * time.Second
}
