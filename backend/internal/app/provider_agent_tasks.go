package app

import "github.com/TokenFlux/TokenRouter/internal/provider"

// provideAgentTaskCoordinator 让所有持久提供商任务入口共享同一进程内按提供商锁。
func provideAgentTaskCoordinator() *provider.OpenAITaskCoordinator {
	return &provider.OpenAITaskCoordinator{}
}
