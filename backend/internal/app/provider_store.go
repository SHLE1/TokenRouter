package app

import (
	"database/sql"
	"time"

	dbent "github.com/TokenFlux/TokenRouter/ent"
	billingpostgres "github.com/TokenFlux/TokenRouter/internal/billing/postgres"
	egresspostgres "github.com/TokenFlux/TokenRouter/internal/egress/postgres"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	acctcore "github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	"github.com/TokenFlux/TokenRouter/internal/routing/accessview"
	routingpostgres "github.com/TokenFlux/TokenRouter/internal/routing/postgres"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// provideProviderStore 构造共享的提供商存储，并绑定数据转换和事件写入函数。
func provideProviderStore(client *dbent.Client, db *sql.DB, cache scheduler.SnapshotCache) *providerpostgres.ProviderStore {
	store := providerpostgres.NewProviderStore(client, db, providerpostgres.ProviderStoreOptions{
		Group: func(g *dbent.Group) *accessview.GroupConfig {
			return (*accessview.GroupConfig)(routingpostgres.GroupFromEnt(g))
		},
		OllamaIdentity: acctcore.IsOllamaCloudUsageProvider,
		Proxy:          egresspostgres.ProxyEntity, Now: time.Now, LoadLocation: time.LoadLocation,
		Observe: func(format string, args ...any) { logging.LegacyPrintf("repository.provider", format, args...) },
	})
	store.SetEvents(newProviderEvents(store, cache))
	return store
}

func provideExecutionProviderStore(store *providerpostgres.ProviderStore, usage *billingpostgres.ProviderUsageStore) gatewayprovider.ExecutionProviderStore {
	return &executionProviderStore{data: store, usage: usage}
}
