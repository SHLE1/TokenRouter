package app

import (
	egressprovider "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideQoderRequestRefresh 为 Qoder 入站请求绑定共享的刷新协调器、提供商存储和会话缓存。
func provideQoderRequestRefresh(store *providerpostgres.ProviderStore, tokens *provideradapter.QoderTokenProvider, coordinator *provider.OAuthRefreshAPI, transport provideradapter.QoderTransport, profiles *egressprovider.TLSProfiles) *provideradapter.QoderRequestRefresh {
	return &provideradapter.QoderRequestRefresh{Store: store, Tokens: tokens, Coordinator: coordinator, Transport: transport, Profiles: profiles}
}

// provideQoderRuntime 绑定共享的 token 源、传输池和提供商存储，Qoder 执行组件管理会话与执行器。
func provideQoderRuntime(tokens *provideradapter.QoderTokenProvider, transport provideradapter.QoderTransport, profiles *egressprovider.TLSProfiles, store *providerpostgres.ProviderStore) *gatewayprovider.QoderRuntime {
	return gatewayprovider.NewQoderRuntime(gatewayprovider.QoderRuntimeOptions{Tokens: tokens, Transport: transport, Profiles: profiles, Health: store})
}
