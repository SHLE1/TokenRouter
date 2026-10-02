package app

import (
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/gateway"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressprovider "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideProviderModelSync 为模型目录同步绑定提供商、凭据和请求函数，并管理共享查询实例的启停。
func provideProviderModelSync(store *postgres.ProviderStore, transport httpclient.UpstreamTransport, claude *provider.ClaudeTokenSource, gemini *provider.GeminiTokenSource, grok *provider.GrokTokenSource, antigravity *provider.AntigravityTokenSource, profiles *egressprovider.TLSProfiles, tasks *provideradapter.ProbeTasks, settings *gateway.RuntimeSettings, cfg *config.Config, manager *lifecycle.Manager) *provider.ModelSyncService {
	policy := egress.OperatorURLPolicy{Enabled: cfg.Security.URLAllowlist.Enabled, AllowInsecureHTTP: cfg.Security.URLAllowlist.AllowInsecureHTTP, AllowPrivateHosts: cfg.Security.URLAllowlist.AllowPrivateHosts, UpstreamHosts: slices.Clone(cfg.Security.URLAllowlist.UpstreamHosts)}
	limit := resolveModelsListReadLimit(cfg)
	queries := &provideradapter.ModelCatalogue{
		Transport: transport, Profiles: profiles, ClaudeTokens: claude, GeminiTokens: gemini, GrokTokens: grok, AntigravityTokens: antigravity, DefaultGrokBaseURL: gatewayprovider.GrokDefaultBaseURLReader(settings), Read: store.GetByID, EnsureTask: tasks.Ensure,
		Options: provideradapter.ModelCatalogueOptions{ValidateURL: policy.Validate, OperatorValidator: policy.Validate, BodyLimit: limit, CodexModelsURL: provideradapter.DefaultCodexModelsURL},
	}
	core := provider.NewModelSyncService(queries.FetchUpstreamSupportedModels)
	manager.Register(lifecycle.Hook{Name: "ProviderModelList", StopOrder: 26, Stop: core.StopContext})
	return core
}
