package app

import provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"

// provideUpstreamHealth 返回 app 已构造的共享健康状态组件。
func provideUpstreamHealth(runtime *providerHealthRuntime) *provideradapter.UpstreamHealth {
	return runtime.Observer
}
