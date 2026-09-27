//go:build unit

package googleforward_test

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/googleforward"
	gatewaytestkit "github.com/TokenFlux/TokenRouter/internal/gateway/testkit"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// newUpstreamHealthForTest 仅投影原测试配置，不保留旧服务或调用代理。
func newUpstreamHealthForTest(store gatewayprovider.ExecutionProviderStore, _ *googleforward.Options, cache provider.TempUnschedCache, options provider.HealthOptions, readers *gatewayprovider.RuntimeReaders) *provideradapter.UpstreamHealth {
	return gatewaytestkit.NewHealthObserver(gatewaytestkit.HealthInput{Store: store, Cache: cache, Options: options, Readers: readers})
}
