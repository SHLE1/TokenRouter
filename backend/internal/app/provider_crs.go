package app

import (
	"log/slog"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/config"
	egresspostgres "github.com/TokenFlux/TokenRouter/internal/egress/postgres"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerhttp "github.com/TokenFlux/TokenRouter/internal/provider/httpapi"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideCRSSync 为 CRS 同步绑定共享的提供商存储、代理存储和刷新协调器。
func provideCRSSync(providers *providerpostgres.ProviderStore, proxies *egresspostgres.ProxyStore, oauth *provider.ClaudeAuthorization, openai *provider.OpenAIAuthorization, gemini *provider.GeminiAuthorization, cfg *config.Config, coordinator *provider.OAuthRefreshAPI) *provider.CRSSync {
	exchange := &provider.CRSAuthorization{Claude: oauth, OpenAI: openai, Gemini: gemini}
	v := cfg.Security.URLAllowlist
	client := provideradapter.NewCRSClient(provideradapter.CRSClientOptions{Configured: true, AllowlistEnabled: v.Enabled, AllowInsecureHTTP: v.AllowInsecureHTTP, AllowPrivateHosts: v.AllowPrivateHosts, Hosts: v.CRSHosts})
	return provider.NewCRSSync(providers, proxies, client, provider.CRSOptions{Now: time.Now, Warn: slog.Warn, Refresh: exchange.Coordinated(coordinator, provideradapter.GeminiTokenCacheKey)})
}

func provideCRSHTTP(source *provider.CRSSync) *providerhttp.CRSHandler {
	return providerhttp.NewCRSHandler(source)
}
