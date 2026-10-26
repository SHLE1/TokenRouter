package provider

import (
	"strings"
)

// base_url.go selects OpenAI-compatible request addresses from
// protocol-specific configuration, legacy credentials, and platform defaults.
// Kimi, Zhipu, and DeepSeek providers can configure a Chat-specific address
// independently of the legacy base URL. The presence of unified protocol
// settings or adaptive routing selects that address when populated. Saved
// legacy addresses apply to API-key and upstream record types, and remaining
// choices use platform and service-mode defaults.
//
// Address selection keeps the protocol-specific configuration ahead of its
// legacy counterpart for Kimi, Zhipu, and DeepSeek routes. Kimi and Zhipu
// distinguish coding from payg defaults, while DeepSeek has one default and
// OpenAI uses its official address. Unsupported platforms return an empty
// result so their execution paths can use their own address resolution.

// OpenAIBaseURL resolves a nonnil OpenAI or Kimi, Zhipu, and DeepSeek record's
// upstream address. Unified-key presence or adaptive Kimi, Zhipu, and DeepSeek
// routes prefer a populated Chat address; saved base URLs apply to API-key and
// upstream types before defaults are used.
//
// Kimi, Zhipu, and DeepSeek addresses depend on service mode when no applicable
// saved address exists. The returned address is trimmed but is not validated
// here; execution callers apply their outbound URL policy. Unsupported
// platforms return an empty string, directing callers to their
// platform-specific address helper.
func (r *Record) OpenAIBaseURL(adaptive bool) string {
	if !r.IsOpenAI() && !r.IsKimiZhipuOrDeepSeek() {
		return ""
	}
	if _, unified := r.Credentials[UpstreamProtocolsKey]; r.IsKimiZhipuOrDeepSeek() && (unified || adaptive) {
		if baseURL := r.protocolSlotBaseURL(APIProtocolChatCompletions); baseURL != "" {
			return baseURL
		}
	}
	if r.Type == ProviderTypeAPIKey || r.Type == ProviderTypeUpstream {
		if baseURL := strings.TrimSpace(r.GetCredential("base_url")); baseURL != "" {
			return baseURL
		}
	}
	// No applicable stored address was selected. Kimi and Zhipu mode defaults
	// choose their payg or coding endpoint, while DeepSeek and OpenAI use their platform address.
	switch r.Platform {
	case PlatformKimi:
		if r.GetProviderMode() == ProviderModeCoding {
			return DefaultKimiCodingBaseURL
		}
		return DefaultKimiPayGBaseURL
	case PlatformZhipu:
		if r.GetProviderMode() == ProviderModeCoding {
			return DefaultZhipuCodingBaseURL
		}
		return DefaultZhipuPayGBaseURL
	case PlatformDeepseek:
		return DefaultDeepseekBaseURL
	default:
		return "https://api.openai.com"
	}
}

// protocolSlotBaseURL returns the trimmed address saved in one api_base_urls
// slot. Missing maps, missing slots, non-string values, and blank strings all
// return empty, letting callers fall back to their platform defaults.
func (r *Record) protocolSlotBaseURL(slot string) string {
	slots, _ := r.Credentials["api_base_urls"].(map[string]any)
	value, _ := slots[slot].(string)
	return strings.TrimSpace(value)
}
