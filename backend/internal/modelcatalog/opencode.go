package modelcatalog

import (
	"encoding/json"
	"maps"
	"sync"
)

// opencode.go indexes the protocol declarations of OpenCode model offerings.
// Service routes remain separate from the price and attribute indexes because
// an intermediary's encoding determines its request endpoint independently of
// the first-party metadata used for billing. Each mode retains exact model IDs;
// the same model can use different encodings on payg and Go.
//
// The embedded publication provides deterministic startup defaults. A live
// model synchronization can obtain declarations from another catalog body and
// save them with the provider, allowing its administrator to review route
// changes before they affect inference. Returned maps are detached from the
// retained publication so request-specific overrides cannot mutate defaults.

// embeddedOpenCodeProtocols parses the bundled publication once and retains
// immutable mode tables; callers obtain independent copies through the accessor.
var embeddedOpenCodeProtocols = sync.OnceValue(func() map[string]map[string]string {
	body, err := Offline()
	if err != nil {
		// An unreadable bundled catalog provides no verified routes. Unknown models
		// then require administrator overrides instead of choosing a guessed endpoint.
		return nil
	}
	return ParseOpenCodeProtocols(body)
})

// OpenCodeProtocols returns copied protocol defaults for one service mode.
// Unknown modes return an empty map. Defaults come from the bundled models.dev
// publication and can be replaced by saved provider model declarations.
func OpenCodeProtocols(mode string) map[string]string {
	return maps.Clone(embeddedOpenCodeProtocols()[mode])
}

// ParseOpenCodeProtocols extracts exact model routes from a models.dev catalog.
// Invalid JSON returns no routes. Model SDK declarations override the provider
// SDK, and decision models use SystemOne regardless of an inherited chat SDK.
//
// SDK identities describe wire formats rather than inferred model families.
// Unrecognized encodings remain absent so a future format cannot silently use
// an unrelated endpoint. This parser deliberately excludes prices and model
// attributes; those continue through the catalog's normal validation and
// first-party selection. Callers own every map in the returned result and may
// combine the selected mode with administrator-supplied overrides.
func ParseOpenCodeProtocols(body []byte) map[string]map[string]string {
	var source struct {
		Providers map[string]struct {
			NPM    string `json:"npm"` // provider default SDK
			Models map[string]struct {
				Type     string `json:"type"` // declared model operation
				Provider struct {
					NPM string `json:"npm"`
				} `json:"provider"` // per-model SDK override
			} `json:"models"` // exact model identity records
		} `json:"providers"` // service-specific model offerings
	}
	if json.Unmarshal(body, &source) != nil {
		return nil
	}
	result := make(map[string]map[string]string, 2)
	for mode, name := range map[string]string{"payg": "opencode", "go": "opencode-go"} {
		service := source.Providers[name]
		routes := make(map[string]string)
		for id, model := range service.Models {
			sdk := model.Provider.NPM
			if sdk == "" {
				sdk = service.NPM
			}
			var target string
			switch sdk {
			case "@ai-sdk/openai":
				target = "openai_responses"
			case "@ai-sdk/openai-compatible":
				target = "openai_chat_completions"
			case "@ai-sdk/anthropic":
				target = "anthropic_messages"
			case "@ai-sdk/google":
				target = "gemini_generate_content"
			}
			// Decision entries share a provider-level SDK in older publications, but
			// their declared operation requires the independent SystemOne request shape.
			if model.Type == "decision" {
				target = "systemone"
			}
			if target != "" {
				routes[id] = target
			}
		}
		result[mode] = routes
	}
	return result
}
