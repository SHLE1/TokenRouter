package app

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// provideProviderRuntimeState 构造恢复、刷新和执行入口共享的运行状态。
func provideProviderRuntimeState() *provider.RuntimeBlockState {
	return provider.NewRuntimeBlockState(time.Now)
}
