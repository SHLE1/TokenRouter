package app

import (
	"log"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/infra/timingwheel"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	providerpostgres "github.com/TokenFlux/TokenRouter/internal/provider/postgres"
)

// provideProviderDeferred 为延迟写入绑定提供商存储和时间轮。
func provideProviderDeferred(store *providerpostgres.ProviderStore, wheel *timingwheel.Wheel) *provider.DeferredService {
	return provider.NewDeferredService(store, wheel, provider.DeferredOptions{Interval: 10 * time.Second, Now: time.Now, Observe: log.Printf})
}
