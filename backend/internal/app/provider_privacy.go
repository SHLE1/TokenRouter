package app

import (
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"

	egresspostgres "github.com/TokenFlux/TokenRouter/internal/egress/postgres"
)

// provideProviderPrivacy 绑定管理操作与后台刷新共用的隐私设置组件和平台客户端。
func provideProviderPrivacy(store *providerpostgres.ProviderStore, proxies *egresspostgres.ProxyStore, factory openai.PrivacyClientFactory, manager *lifecycle.Manager) *provider.PrivacyService {
	core := provider.NewPrivacyService(store, proxies, provideradapter.PrivacyOptions(factory, openai.PrivacyEndpoints{}))
	manager.Register(lifecycle.Hook{Name: "ProviderPrivacy", StopOrder: 26, Stop: core.StopContext})
	return core
}
