package app

import (
	"log"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
)

// provideProviderExpiry 为每分钟执行的过期检查注入提供商存储，并登记维护任务的启停。
func provideProviderExpiry(store *providerpostgres.ProviderStore) *provider.ExpiryService {
	return provider.NewExpiryService(store, provider.ExpiryOptions{Interval: time.Minute, Now: time.Now, Observe: log.Printf})
}
