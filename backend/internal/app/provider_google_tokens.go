package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideGeminiTokens 组合 project 查询、Vertex 凭据交换和凭据字段保存。
func provideGeminiTokens(store *postgres.ProviderStore, cache provider.AccessTokenCache, authorization *provider.GeminiAuthorization, refresh *provider.OAuthRefreshAPI) *provider.GeminiTokenSource {
	executor := &provider.GeminiTokenRefresher{Authorization: authorization, Key: provideradapter.GeminiTokenCacheKey}
	return &provider.GeminiTokenSource{Options: provider.GeminiTokenOptions{
		Cache: cache, Repository: store, Policy: provider.GeminiProviderRefreshPolicy(),
		Debug: slog.Debug, Warn: slog.Warn, Logf: log.Printf,
		Project: authorization.FetchProject, ResolveProxy: authorization.Options.ResolveProxy,
		Vertex: func(ctx context.Context, value *provider.Record) (string, error) {
			return provideradapter.VertexServiceAccountAccessToken(ctx, cache, value)
		},
		Persist: func(ctx context.Context, value *provider.Record, credentials map[string]any) error {
			_, err := provider.PersistCredentials(ctx, store, value, credentials, slog.Warn)
			return err
		},
		Refresh: func(ctx context.Context, value *provider.Record, window time.Duration) (*provider.OAuthRefreshResult, error) {
			return refresh.RefreshIfNeeded(ctx, value, executor, window)
		},
	}}
}

// provideAntigravityTokens 绑定共享的回填状态、冷却缓存和刷新协调器。
func provideAntigravityTokens(store *postgres.ProviderStore, cache provider.AccessTokenCache, authorization *provider.AntigravityAuthorization, refresh *provider.OAuthRefreshAPI, cooldown provider.TempUnschedCache) *provider.AntigravityTokenSource {
	executor := &provider.AntigravityRefreshRules{
		RefreshProviderToken:     authorization.RefreshProviderToken,
		BuildProviderCredentials: authorization.BuildProviderCredentials,
		Printf:                   func(format string, args ...any) { _, _ = fmt.Printf(format, args...) },
		Logf:                     log.Printf,
	}
	return &provider.AntigravityTokenSource{Options: provider.AntigravityTokenOptions{
		Cache: cache, Repository: store, Policy: provider.AntigravityProviderRefreshPolicy(),
		Debug: slog.Debug, Warn: slog.Warn, TempUnschedCache: cooldown,
		SetTempUnschedulable: store.SetTempUnschedulable,
		FillProject:          authorization.FillProjectID,
		Persist: func(ctx context.Context, value *provider.Record, credentials map[string]any) error {
			_, err := provider.PersistCredentials(ctx, store, value, credentials, slog.Warn)
			return err
		},
		Refresh: func(ctx context.Context, value *provider.Record, window time.Duration) (*provider.OAuthRefreshResult, error) {
			return refresh.RefreshIfNeeded(ctx, value, executor, window)
		},
	}}
}
