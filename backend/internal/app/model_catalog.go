package app

import (
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/modelidentity"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"
)

// provideModelCatalogService 从 bootstrap 配置提取模型目录服务的参数。
// ModelCatalogInitialization 和 ModelCatalogService hook 分别负责初始化及周期更新的启停。
func provideModelCatalogService(cfg *config.Config, remote provider.RemoteClient) (*provider.Service, error) {
	options := provider.Options{
		DataDir:               cfg.Pricing.DataDir,
		RemoteURL:             cfg.Pricing.RemoteURL,
		FallbackFile:          cfg.Pricing.FallbackFile,
		CheckIntervalMinutes:  cfg.Pricing.CheckIntervalMinutes,
		URLAllowlistEnabled:   cfg.Security.URLAllowlist.Enabled,
		AllowInsecureHTTP:     cfg.Security.URLAllowlist.AllowInsecureHTTP,
		AllowPrivateHosts:     cfg.Security.URLAllowlist.AllowPrivateHosts,
		PricingHosts:          slices.Clone(cfg.Security.URLAllowlist.PricingHosts),
		ModelLookupCandidates: modelidentity.CandidatesFactory,
	}
	return provider.NewService(options, remote), nil
}

// provideModelCatalogRemoteClient 配置更新代理和直连回退开关。
func provideModelCatalogRemoteClient(cfg *config.Config) provider.RemoteClient {
	return provider.NewRemoteClient(cfg.Update.ProxyURL, cfg.Security.ProxyFallback.AllowDirectOnError)
}
