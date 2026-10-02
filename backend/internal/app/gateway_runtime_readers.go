package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/moderation"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/search"
	"github.com/TokenFlux/TokenRouter/internal/settings"
	"github.com/TokenFlux/TokenRouter/internal/upstream/antigravity"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// provideGatewayRuntimeReaders 绑定共享读取器，动态设置在请求期间读取。
func provideGatewayRuntimeReaders(store *settings.Store, gatewayRuntime *gateway.RuntimeSettings, providerRuntime *provider.RuntimeSettings, quota *provider.QuotaSettingsCache, routingRuntime *routing.RuntimeSettings, moderationRuntime *moderation.RuntimeSettings, searchRuntime *search.ConfigService) *gatewayprovider.RuntimeReaders {
	antigravity.SetUserAgentVersionResolver(gatewayRuntime.GetAntigravityUserAgentVersion)
	openai.SetCodexCanonicalUserAgentResolver(func() string { return gatewayRuntime.GetOpenAICodexUserAgent(context.Background()) })
	readers := &gatewayprovider.RuntimeReaders{Gateway: gatewayRuntime, Provider: providerRuntime, Quota: quota, Routing: routingRuntime, Moderation: moderationRuntime, Search: searchRuntime, Scheduler: store}
	return readers
}

// provideModerationSettings 为网关审核绑定独立的设置缓存。
func provideModerationSettings(store *settings.Store) *moderation.RuntimeSettings {
	return moderation.NewRuntimeSettings(store, settings.ErrSettingNotFound)
}
