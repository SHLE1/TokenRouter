package app

import (
	"context"
	"log/slog"
	"time"

	egressadapter "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"

	billingpostgres "github.com/TokenFlux/TokenRouter/internal/billing/postgres"

	egresspostgres "github.com/TokenFlux/TokenRouter/internal/egress/postgres"
	"github.com/TokenFlux/TokenRouter/internal/routing"

	routingpostgres "github.com/TokenFlux/TokenRouter/internal/routing/postgres"
	"github.com/google/uuid"
)

// provideProviderAdmin 绑定唯一管理用例与原平台执行端口，构造不运行后台任务。
func provideProviderAdmin(store *providerpostgres.ProviderStore, usage *billingpostgres.ProviderUsageStore, blocker provider.RuntimeUnblocker, privacy *provider.PrivacyService, groups *routingpostgres.GroupStore, proxies *egresspostgres.ProxyStore, tasks *lifecycle.Tasks, upstream httpclient.UpstreamTransport, tls *egressadapter.TLSProfiles) *provider.Admin {
	return provider.NewAdmin(store, provider.AdminOptions{ShadowModels: provideradapter.DefaultSparkShadowModels, Duplicates: store, Quotas: usage, RuntimeBlocker: blocker, Privacy: privacy, Groups: providerGroupReferences{groups}, Proxies: proxies, Creation: provider.CreationOptions{Now: time.Now, LoadLocation: time.LoadLocation, NewSeed: uuid.NewString}, Credentials: provideradapter.CreateCredentialHooks(upstream, tls), Background: tasks.Go, Error: slog.Error})
}

type providerGroupReferences struct{ store *routingpostgres.GroupStore }

func (g providerGroupReferences) GetGroup(ctx context.Context, id int64) (*provider.GroupReference, error) {
	value, err := g.store.GetByID(ctx, id)
	if value == nil {
		return nil, err
	}
	return &provider.GroupReference{ID: value.ID, Name: value.Name, RequireOAuthOnly: value.RequireOAuthOnly}, err
}

func (g providerGroupReferences) ActiveGroups(ctx context.Context, platform string) ([]provider.GroupReference, error) {
	rows, err := g.store.ListActive(ctx)
	if rows == nil {
		return nil, err
	}
	out := make([]provider.GroupReference, len(rows))
	for i, v := range rows {
		out[i] = provider.GroupReference{ID: v.ID, Name: v.Name, RequireOAuthOnly: v.RequireOAuthOnly}
	}
	return out, err
}

func (g providerGroupReferences) ValidateGroups(ctx context.Context, ids []int64) error {
	return routing.ValidateGroupIDs(ctx, g.store, ids)
}
