package app

import (
	"database/sql"

	billingpostgres "github.com/TokenFlux/TokenRouter/internal/billing/postgres"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// provideProviderUsage 为管理和用量累计入口绑定共享的资金存储。
func provideProviderUsage(db *sql.DB, store *providerpostgres.ProviderStore, cache scheduler.SnapshotCache) *billingpostgres.ProviderUsageStore {
	return billingpostgres.NewProviderUsageStore(db, providerUsageEvents(store, cache, db))
}
