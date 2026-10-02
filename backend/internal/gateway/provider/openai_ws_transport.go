package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// ResolveOpenAIWSTransport 将提供商资格传给 egress，取得 WebSocket 传输选择结果。
func ResolveOpenAIWSTransport(value *provider.Record, options *egress.OpenAIWSOptions, defaultMode string) egress.OpenAIWSProtocolDecision {
	input := egress.OpenAIWSProvider{}
	if value != nil {
		input = egress.OpenAIWSProvider{
			Present:     true,
			OpenAI:      value.IsOpenAI(),
			ForceHTTP:   value.IsOpenAIWSForceHTTPEnabled(),
			OAuthLike:   value.IsOpenAIOAuthLike(),
			APIKey:      value.IsOpenAIApiKey(),
			WSEnabled:   value.IsOpenAIResponsesWebSocketV2Enabled(),
			Concurrency: value.Concurrency,
		}
		if options != nil && options.ModeRouterV2Enabled {
			input.Mode = value.ResolveOpenAIResponsesWebSocketV2Mode(defaultMode)
		}
	}
	return egress.ResolveOpenAIWSTransport(input, options)
}
