package app

import (
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/transport"
)

// provideHTTPUpstream 构造共享传输实例，每次调用时读取传输所需的配置。
// @project-doc docs/architecture/system_architecture.md#dependency_layers
func provideHTTPUpstream(cfg *config.Config) *transport.Client {
	return transport.New(func() *transport.Options {
		return httpTransportOptions(cfg)
	})
}

func httpTransportOptions(cfg *config.Config) *transport.Options {
	if cfg == nil {
		return nil
	}
	gateway := cfg.Gateway
	http2 := gateway.OpenAIHTTP2
	return &transport.Options{
		ValidateResolvedIP:          cfg.Security.URLAllowlist.Enabled && !cfg.Security.URLAllowlist.AllowPrivateHosts,
		ConnectionPoolIsolation:     gateway.ConnectionPoolIsolation,
		MaxUpstreamClients:          gateway.MaxUpstreamClients,
		ClientIdleTTLSeconds:        gateway.ClientIdleTTLSeconds,
		MaxIdleConns:                gateway.MaxIdleConns,
		MaxIdleConnsPerHost:         gateway.MaxIdleConnsPerHost,
		MaxConnsPerHost:             gateway.MaxConnsPerHost,
		IdleConnTimeoutSeconds:      gateway.IdleConnTimeoutSeconds,
		ResponseHeaderTimeout:       gateway.ResponseHeaderTimeout,
		OpenAIResponseHeaderTimeout: gateway.OpenAIResponseHeaderTimeout,
		GrokResponseHeaderTimeout:   gateway.GrokResponseHeaderTimeout,
		OpenAIHTTP2: transport.HTTP2Options{
			Enabled:                   http2.Enabled,
			AllowProxyFallbackToHTTP1: http2.AllowProxyFallbackToHTTP1,
			FallbackErrorThreshold:    http2.FallbackErrorThreshold,
			FallbackWindowSeconds:     http2.FallbackWindowSeconds,
			FallbackTTLSeconds:        http2.FallbackTTLSeconds,
		},
	}
}
