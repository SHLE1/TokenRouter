package text

import (
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"

	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
)

// ResponsesCapability 根据请求的生图意图返回所需端点能力。
func ResponsesCapability(imageIntent bool, platform string) providercore.OpenAIEndpointCapability {
	if imageIntent && platform == capability.PlatformOpenAI {
		return providercore.OpenAIEndpointCapabilityResponses
	}
	return providercore.OpenAIEndpointCapabilityTextGeneration
}

// RequiredResponsesCapability 为旧版压缩返回 Responses 能力，为 V2 返回 RemoteCompactionV2 能力。
func RequiredResponsesCapability(imageIntent bool, nativeCompactionV2 bool, legacyCompact bool, platform string) providercore.OpenAIEndpointCapability {
	if nativeCompactionV2 && platform == capability.PlatformOpenAI {
		return providercore.OpenAIEndpointCapabilityRemoteCompactionV2
	}
	if legacyCompact && platform == capability.PlatformOpenAI {
		return providercore.OpenAIEndpointCapabilityResponses
	}
	return ResponsesCapability(imageIntent, platform)
}
