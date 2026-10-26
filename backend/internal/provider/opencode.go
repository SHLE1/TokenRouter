package provider

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/modelcatalog"
	infraerrors "github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/protocol"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// opencode.go resolves OpenCode service modes, model encodings, and endpoints.
// A provider carries one API key and two independent sets of optional model
// overrides, indexed by payg and Go. Encoding selection uses the final mapped
// model and exact catalog identity, preventing an incoming client protocol
// from choosing an unsupported endpoint for that model.
//
// Protocol enablement remains an administrator restriction over these routes.
// Address slots replace complete protocol bases, while base_url replaces the
// inference service root used for defaults and model listing. These functions
// select addresses without contacting upstream; transport callers validate
// every selected URL with the configured egress policy before sending secrets.

// OpenCode configuration keys identify the subscription mode and mode-scoped
// exact-model protocol overrides persisted with the other provider credentials.
const (
	ProviderModeGo            = "go"                       // OpenCode subscription service mode
	OpenCodeModelProtocolsKey = "opencode_model_protocols" // exact model override key
)

// IsOpenCode reports whether this record uses the OpenCode inference platform.
func (r *Record) IsOpenCode() bool { return r != nil && r.Platform == PlatformOpenCode }

// OpenCodeMode returns the configured mode or the default payg mode.
// Validation rejects other saved values; read paths retain invalid nonempty
// values so they cannot acquire another mode's catalog routes accidentally.
func (r *Record) OpenCodeMode() string {
	mode := strings.TrimSpace(r.GetCredential("provider_mode"))
	if mode == "" {
		return ProviderModePayG
	}
	return mode
}

// OpenCodeModelProtocols returns mode-specific catalog routes with exact saved
// overrides applied. The result is independently owned, and non-OpenCode records
// return nil so ordinary provider routing retains its existing protocol rules.
//
// Only the selected mode's overrides participate: a model's payg encoding can
// differ from its Go encoding. Defaults use the bundled publication, while an
// administrator can save a synchronized or manually selected encoding. Unknown
// model IDs stay absent and therefore cannot acquire a route merely because the
// provider has enabled a compatible protocol family.
// @project-doc docs/interfaces/opencode_upstream.md#opencode_routing
func (r *Record) OpenCodeModelProtocols() map[string]capability.ProtocolID {
	if !r.IsOpenCode() {
		return nil
	}
	mode := r.OpenCodeMode()
	result := make(map[string]capability.ProtocolID)
	for id, value := range modelcatalog.OpenCodeProtocols(mode) {
		result[id] = capability.ProtocolID(value)
	}
	modes, _ := r.Credentials[OpenCodeModelProtocolsKey].(map[string]any)
	overrides, _ := modes[mode].(map[string]any)
	for id, value := range overrides {
		if text, ok := value.(string); ok {
			result[id] = capability.ProtocolID(text)
		}
	}
	return result
}

// OpenCodeRoot returns the custom or official service root for this mode.
// The root excludes protocol-family and version segments. URL validation belongs
// to callers that send inference or model-catalog requests.
func (r *Record) OpenCodeRoot() string {
	if base := strings.TrimSpace(r.GetCredential("base_url")); base != "" {
		return strings.TrimRight(base, "/")
	}
	if r.OpenCodeMode() == ProviderModeGo {
		return "https://opencode.ai/inference/go"
	}
	return "https://opencode.ai/inference"
}

// OpenCodeBaseURL resolves a complete protocol base including its version.
// A populated address slot overrides the service root for that protocol. An
// unsupported protocol returns empty and cannot produce an inference endpoint.
func (r *Record) OpenCodeBaseURL(target protocol.ProtocolID) string {
	var slot, suffix string
	switch target {
	case protocol.ProtocolAnthropicMessages:
		slot, suffix = "anthropic", "/anthropic/v1"
	case protocol.ProtocolOpenAIResponses:
		slot, suffix = "responses", "/openai/v1"
	case protocol.ProtocolOpenAIChatCompletions:
		slot, suffix = "chat_completions", "/openai/v1"
	case protocol.ProtocolGeminiGenerateContent:
		slot, suffix = "gemini", "/google/v1beta"
	case protocol.ProtocolSystemOne:
		slot, suffix = "systemone", "/systemone/v1"
	default:
		return ""
	}
	bases, _ := r.Credentials["api_base_urls"].(map[string]any)
	if custom, ok := bases[slot].(string); ok && strings.TrimSpace(custom) != "" {
		return strings.TrimRight(strings.TrimSpace(custom), "/")
	}
	return r.OpenCodeRoot() + suffix
}

// ValidateOpenCodeCredentials checks authentication, mode, model protocols,
// and address shapes before persistence. Missing mode means payg, while explicit
// invalid types or unknown protocols fail before any record can be stored.
func ValidateOpenCodeCredentials(r *Record) error {
	if !r.IsOpenCode() {
		return nil
	}
	// invalid gives administrator input failures the shared OpenCode configuration reason.
	invalid := func(message string) error { return infraerrors.BadRequest("OPENCODE_CONFIG_INVALID", message) }
	if r.Type != ProviderTypeAPIKey {
		return invalid("OpenCode requires API Key credentials")
	}
	if strings.TrimSpace(r.GetCredential("api_key")) == "" {
		return invalid("OpenCode API key is required")
	}
	if raw, exists := r.Credentials["provider_mode"]; exists {
		mode, ok := raw.(string)
		if !ok || (mode != ProviderModePayG && mode != ProviderModeGo) {
			return invalid("OpenCode provider_mode must be payg or go")
		}
	}
	if raw, exists := r.Credentials[OpenCodeModelProtocolsKey]; exists {
		modes, ok := raw.(map[string]any)
		if !ok {
			return invalid("opencode_model_protocols must be an object keyed by payg or go")
		}
		for mode, rawModels := range modes {
			if mode != ProviderModePayG && mode != ProviderModeGo {
				return invalid("Unknown OpenCode model protocol mode")
			}
			models, ok := rawModels.(map[string]any)
			if !ok {
				return invalid("OpenCode model protocols must be an object")
			}
			for id, rawProtocol := range models {
				value, ok := rawProtocol.(string)
				if strings.TrimSpace(id) == "" || id != strings.TrimSpace(id) || !ok || !slices.Contains(r.NativeProtocolOptions(), capability.ProtocolID(value)) {
					return invalid(fmt.Sprintf("Invalid OpenCode protocol for model %q", id))
				}
			}
		}
	}
	addresses := []string{}
	if raw, exists := r.Credentials["base_url"]; exists {
		base, ok := raw.(string)
		if !ok {
			return invalid("OpenCode base_url must be a string")
		}
		if strings.TrimSpace(base) != "" {
			addresses = append(addresses, base)
		}
	}
	if raw, exists := r.Credentials["api_base_urls"]; exists {
		bases, ok := raw.(map[string]any)
		if !ok {
			return invalid("api_base_urls must be an object")
		}
		for slot, rawBase := range bases {
			base, ok := rawBase.(string)
			if !ok || !slices.Contains([]string{"anthropic", "responses", "chat_completions", "gemini", "systemone"}, slot) {
				return invalid("Invalid OpenCode protocol address")
			}
			if strings.TrimSpace(base) != "" {
				addresses = append(addresses, base)
			}
		}
	}
	for _, base := range addresses {
		parsed, err := url.Parse(strings.TrimSpace(base))
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return invalid("Invalid OpenCode base URL")
		}
	}
	return nil
}
