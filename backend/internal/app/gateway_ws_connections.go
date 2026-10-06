package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway/ws"
	openaiws "github.com/TokenFlux/TokenRouter/internal/upstream/openai/ws"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// provideWSConnections 共享平台拨号器并保留连接池的按需启动时点。
func provideWSConnections(cfg *config.Config, manager *lifecycle.Manager, runtime *ws.Runtime) *openaiws.OpenAIWSConnections {
	connections := openaiws.NewOpenAIWSConnections(openAIWSPoolOptions(cfg), openai.NewDefaultWSClientDialer())
	runtime.SetApply(func(value ws.Parameters) error { return connections.UpdateOptions(responsesWSPoolOptions(value)) })
	manager.Register(lifecycle.Hook{Name: "OpenAIWSConnections", StartOrder: 990, StopOrder: 10, Stop: func(context.Context) error { connections.Close(); return nil }})
	return connections
}

func openAIWSPoolOptions(cfg *config.Config) *openaiws.WSPoolOptions {
	if cfg == nil {
		return nil
	}
	returnValue := openAIWSExecutionOptions(cfg)
	returnPtr := responsesWSPoolOptions(*returnValue)
	return &returnPtr
}

// responsesWSPoolOptions 提取连接池使用的参数。
func responsesWSPoolOptions(options ws.Parameters) openaiws.WSPoolOptions {
	return openaiws.WSPoolOptions{
		MaxConnsPerProvider:                         options.MaxConnsPerProvider,
		DynamicMaxConnsByProviderConcurrencyEnabled: options.DynamicMaxConnsByProviderConcurrencyEnabled,
		OAuthMaxConnsFactor:                         options.OAuthMaxConnsFactor,
		APIKeyMaxConnsFactor:                        options.APIKeyMaxConnsFactor,
		MinIdlePerProvider:                          options.MinIdlePerProvider,
		MaxIdlePerProvider:                          options.MaxIdlePerProvider,
		QueueLimitPerConn:                           options.QueueLimitPerConn,
		PoolTargetUtilization:                       options.PoolTargetUtilization,
		PrewarmCooldownMS:                           options.PrewarmCooldownMS,
		DialTimeoutSeconds:                          options.DialTimeoutSeconds,
	}
}
