package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideClaudeTokens 绑定 OAuth 与 Vertex 凭据源，各来源按调用时机使用缓存。
func provideClaudeTokens(store *postgres.ProviderStore, cache provider.AccessTokenCache, authorization *provider.ClaudeAuthorization, refresh *provider.OAuthRefreshAPI) *provider.ClaudeTokenSource {
	executor := &provider.ClaudeTokenRefresher{Authorization: authorization}
	return &provider.ClaudeTokenSource{Options: provider.ClaudeTokenOptions{
		Cache: cache, Repository: store, Policy: provider.ClaudeProviderRefreshPolicy(),
		Debug: slog.Debug, Warn: slog.Warn,
		Vertex: func(ctx context.Context, value *provider.Record) (string, error) {
			return provideradapter.VertexServiceAccountAccessToken(ctx, cache, value)
		},
		Refresh: func(ctx context.Context, value *provider.Record, window time.Duration) (*provider.OAuthRefreshResult, error) {
			return refresh.RefreshIfNeeded(ctx, value, executor, window)
		},
	}}
}

// provideMessageCredentials 复用 Claude 和 Vertex 凭据源，其他平台读取已保存的凭据。
func provideMessageCredentials(claude *provider.ClaudeTokenSource) *provider.MessageCredentialSource {
	result := &provider.MessageCredentialSource{}
	if claude != nil {
		result.Claude = claude.GetAccessToken
	}
	return result
}
