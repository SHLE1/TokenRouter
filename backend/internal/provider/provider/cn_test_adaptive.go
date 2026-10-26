package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/protocol"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	claude "github.com/TokenFlux/TokenRouter/internal/upstream/anthropic"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// cn_test_adaptive.go tests enabled CN native protocols in sequence,
// optionally narrowed to one selected protocol. An empty selection checks
// enabled Chat, Messages, and Responses endpoints in that order, ending at the
// first failed endpoint.
//
// Each endpoint can emit content and status events, but completion is
// suppressed until every selected endpoint succeeds. Chat owns its start
// event; tests without Chat commit one start before the other endpoints run.
// Model preparation applies one mapping for Messages and Responses, while Chat
// delegates its original supplied model to the fixed executor. Messages and
// Responses run through the native exchange shared with OpenCode probes, which
// sends stateless CN Responses and attempts provider error persistence after
// an authentication failure.

// ExecuteAdaptive tests one selected native protocol or every enabled protocol
// for a CN record. Empty protocol sets and disabled explicit selections fail
// before an upstream request; a blank Messages or Responses model uses the
// OpenAI test default before mapping.
//
// Chat, Messages, and Responses execute in a fixed order and stop on their
// first failure. Completion suppression spans those attempts so intermediate
// parsers cannot declare the entire test successful. Chat supplies its own
// start event; otherwise this coordinator emits one before execution. The
// suppression flag is cleared on return and before the single final success
// event.
func (s *CNProviderTest) ExecuteAdaptive(c *TestRun, value *providercore.Record, modelID string, prompt string, selected protocol.ProtocolID) error {
	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = openai.DefaultTestModel
	}
	testModelID = mappedTestModel(value, testModelID)
	if c.ProtocolRoute.ClientProtocol != "" {
		if _, err := resolveProbeType(c.Context, s.Attributes, value, c.GroupID, testModelID, string(selected), providercore.ProviderTestTypeText); err != nil {
			return (TestStreamOutput{}).Error(c, err.Error())
		}
	}

	authToken := strings.TrimSpace(value.GetOpenAIProtocolAPIKey())
	if authToken == "" {
		return (TestStreamOutput{}).Error(c, "No API key available")
	}

	// Each endpoint parser can emit success independently. Suppress those
	// intermediate completions so a later endpoint failure cannot follow an
	// already published whole-test success.
	c.SetSuppressCompletion(true)
	defer c.SetSuppressCompletion(false)
	enabled := value.UpstreamProtocols()
	if len(enabled) == 0 {
		return (TestStreamOutput{}).Error(c, "No upstream protocols enabled")
	}
	if selected != "" {
		if !slices.Contains(enabled, selected) {
			return (TestStreamOutput{}).Error(c, fmt.Sprintf("Upstream protocol %s is not enabled", selected))
		}
		enabled = []protocol.ProtocolID{selected}
	}
	c.Begin(false)
	// The fixed Chat executor commits its own start event. When Chat is absent,
	// commit one here before Messages or Responses content can reach the sink.
	if !slices.Contains(enabled, protocol.ProtocolOpenAIChatCompletions) {
		c.Begin(true)
		(TestStreamOutput{}).SendEvent(c, providercore.TestEvent{Type: "test_start", Model: testModelID})
	}
	if slices.Contains(enabled, protocol.ProtocolOpenAIChatCompletions) {
		if err := s.executeChat(c, value, modelID, prompt); err != nil {
			return err
		}
	}

	for _, wire := range []protocol.ProtocolID{protocol.ProtocolAnthropicMessages, protocol.ProtocolOpenAIResponses} {
		if !slices.Contains(enabled, wire) {
			continue
		}
		if err := s.executeAdaptiveEndpoint(c, value, testModelID, prompt, authToken, wire); err != nil {
			return err
		}
	}

	c.SetSuppressCompletion(false)
	(TestStreamOutput{}).SendEvent(c, providercore.TestEvent{Type: "test_complete", Success: true})
	return nil
}

// executeAdaptiveEndpoint runs one mapped native Messages or Responses probe
// with the supplied key and prompt through the shared native exchange. It
// validates that protocol's CN root, then brackets the exchange with testing
// and verified status events so adaptive output names each endpoint.
//
// Messages sends Claude default headers, the API-key beta, and the record's
// API-key scheme; Responses sends a Bearer key with Codex probe identity. Both
// apply allowed header overrides last. The shared exchange closes the response
// and attempts a provider error-state write after a 401.
func (s *CNProviderTest) executeAdaptiveEndpoint(c *TestRun, value *providercore.Record, testModelID, prompt, authToken string, wire protocol.ProtocolID) error {
	name, path, apiProtocol := "Anthropic", "/v1/messages", providercore.APIProtocolAnthropic
	authenticate := func(headers http.Header) {
		for key, text := range claude.DefaultHeaders {
			headers.Set(key, text)
		}
		headers.Set("anthropic-beta", claude.APIKeyBetaHeader)
		claude.SetAPIKeyAuthHeader(headers, value.GetAnthropicAPIKeyAuthScheme() == providercore.AnthropicAPIKeyAuthSchemeAuthorizationBearer, authToken)
	}
	if wire == protocol.ProtocolOpenAIResponses {
		name, path, apiProtocol = "Responses", "/responses", providercore.APIProtocolResponses
		authenticate = func(headers http.Header) {
			headers.Set("Authorization", "Bearer "+authToken)
			openai.ApplyOpenAICodexProbeHeaders(headers)
		}
	}
	baseURL, err := s.ValidateURL((providercore.ProtocolTarget{Record: value}).GetCNProtocolBaseURL(apiProtocol))
	if err != nil {
		return (TestStreamOutput{}).Error(c, fmt.Sprintf("Invalid adaptive %s base URL: %s", name, err.Error()))
	}
	probe := nativeProbe{
		route:   providercore.ResolvedModelRoute{Model: testModelID, Protocol: wire},
		baseURL: baseURL,
		label:   "Adaptive " + name + " endpoint",
		applyHeaders: func(headers http.Header) {
			authenticate(headers)
			applyGrokQuotaHeaders(value, headers)
		},
	}
	body, err := s.probePayload(c, value, probe.route, prompt, providercore.ProviderTestTypeText)
	if err != nil {
		return (TestStreamOutput{}).Error(c, err.Error())
	}
	(TestStreamOutput{}).SendEvent(c, providercore.TestEvent{Type: "status", Text: "Testing adaptive " + name + " endpoint through native " + path})
	if err := s.exchange(c, value, probe, body); err != nil {
		return err
	}
	(TestStreamOutput{}).SendEvent(c, providercore.TestEvent{Type: "status", Text: "Verified through native " + path})
	return nil
}

// cnAnthropicBaseURLMisconfigHint returns a diagnostic for OpenAI-shaped paths
// in an Anthropic root. Invalid or incomplete URLs and unrecognized paths
// return empty text; callers validate the root separately before using this
// hint.
func cnAnthropicBaseURLMisconfigHint(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	path := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	if path == "" {
		return ""
	}
	openAICompatShaped := strings.Contains(path, "/paas/") ||
		strings.HasSuffix(path, "/chat/completions") ||
		strings.HasSuffix(path, "/responses") ||
		httpclient.OpenAIBaseURLHasVersionSuffix(path)
	if !openAICompatShaped {
		return ""
	}
	return fmt.Sprintf(
		"API protocol is anthropic but base_url (%s) looks like an OpenAI-compatible endpoint; set base_url to the provider's Anthropic endpoint (for example https://open.bigmodel.cn/api/anthropic) or switch api_protocol.",
		baseURL,
	)
}
