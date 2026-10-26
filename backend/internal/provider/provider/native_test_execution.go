package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressprovider "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient/tlsfingerprint"
	"github.com/TokenFlux/TokenRouter/internal/protocol"
	"github.com/TokenFlux/TokenRouter/internal/protocol/systemone"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream"
	"github.com/TokenFlux/TokenRouter/internal/upstream/anthropic"
	"github.com/TokenFlux/TokenRouter/internal/upstream/gemini"
	"github.com/TokenFlux/TokenRouter/internal/upstream/jev"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// native_test_execution.go executes probes with an already selected model and
// protocol. The platform target validates its provider configuration and fixes
// the model, endpoint, and authentication callback before handing execution to
// this adapter. Payload construction uses the selected wire protocol, while
// the shared stream readers translate successful replies into test events.
//
// A single TestRun carries group attributes, caller identity, cancellation, and
// sink failures through the exchange. The prepared model passes directly into
// payloads and model-action URLs, so an alias target cannot become another
// mapping input. SystemOne uses its buffered exchange and reports response
// identity and usage validity; streaming protocols reuse their existing readers.

// NativeProviderTest supplies transport and model attributes for prepared probes.
type NativeProviderTest struct {
	// Attributes resolves the final model under the probe's group and selected
	// entry. A nil callback leaves automatic and image output capability unknown.
	Attributes ModelAttributeLookup
	Transport  QoderTransport // provider-aware probe request sender
	// Profiles resolves optional request fingerprints. Nil delegates fingerprint
	// selection to the configured transport using its ordinary default behavior.
	Profiles    *egressprovider.TLSProfiles
	ValidateURL func(string) (string, error) // selected protocol root validator
}

// nativeProbe retains one final route and the platform's request-header policy.
type nativeProbe struct {
	route   provider.ResolvedModelRoute  // resolved model and protocol pair
	request provider.PreparedTestRequest // normalized caller probe input
	baseURL string                       // validated protocol root URL
	// applyHeaders installs the platform's final authentication and caller
	// identity after protocol defaults have initialized the writable header map.
	applyHeaders func(http.Header)
}

// execute sends a prepared probe and parses its protocol's response events.
// The caller owns the TestRun and supplies a final model, validated endpoint,
// and authentication callback. Response bodies close before this method returns.
func (s *NativeProviderTest) execute(run *TestRun, value *provider.Record, probe nativeProbe) error {
	output := TestStreamOutput{}
	kind := probe.request.TestType
	if probe.route.Protocol != protocol.ProtocolGeminiGenerateContent && probe.route.Protocol != protocol.ProtocolSystemOne {
		if kind == provider.ProviderTestTypeImage {
			return output.Error(run, "Selected native protocol supports text probes")
		}
		kind = provider.ProviderTestTypeText
	}
	if probe.route.Protocol != protocol.ProtocolSystemOne {
		var err error
		kind, err = resolveProbeType(run.Context, s.Attributes, value, run.GroupID,
			probe.route.Model, string(probe.route.Protocol), kind)
		if err != nil {
			return output.Error(run, err.Error())
		}
	} else if s.Attributes != nil {
		if _, err := s.Attributes(run.Context, value, run.GroupID, probe.route.Model, string(probe.route.Protocol)); err != nil {
			return output.Error(run, err.Error())
		}
	}
	body, err := s.payload(run, value, probe, kind)
	if err != nil {
		return output.Error(run, err.Error())
	}
	run.Begin(true)
	output.SendEvent(run, provider.TestEvent{Type: "test_start", Model: probe.route.Model})
	// A sink can cancel during Begin or test_start. Observe that cancellation
	// before starting an upstream request whose output cannot reach the caller.
	if err := run.Context.Err(); err != nil {
		return err
	}
	if probe.route.Protocol == protocol.ProtocolSystemOne {
		return executeSystemOneProbe(run, &jev.Target{
			ProviderID: value.ID, URL: jev.EndpointURL(probe.baseURL, "systemone"), Model: probe.route.Model,
			ApplyHeaders: probe.applyHeaders,
			Do:           func(req *http.Request) (*http.Response, error) { return s.send(req, value) },
		}, body)
	}
	endpoint, err := nativeProbeEndpoint(probe)
	if err != nil {
		return output.Error(run, err.Error())
	}
	req, err := http.NewRequestWithContext(run.Context, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return output.Error(run, "Failed to create native probe request: "+err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if probe.route.Protocol == protocol.ProtocolAnthropicMessages {
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	probe.applyHeaders(req.Header)
	resp, err := s.send(req, value)
	if err != nil {
		if run.Context.Err() != nil {
			return run.Context.Err()
		}
		return output.Error(run, "Native probe request failed: "+err.Error())
	}
	// Close successful and rejected responses here because all protocol readers
	// borrow the body, including readers that return after an early terminal event.
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		// Retain any diagnostic prefix returned before a body read failure so the
		// original HTTP rejection remains the primary explanation for this probe.
		body, _ := io.ReadAll(resp.Body)
		return output.Error(run, fmt.Sprintf("Native probe returned %d: %s", resp.StatusCode, body))
	}
	switch probe.route.Protocol {
	case protocol.ProtocolAnthropicMessages:
		if err := output.AdaptiveAnthropic(run, resp.Body); err != nil {
			return err
		}
		output.SendEvent(run, provider.TestEvent{Type: "test_complete", Success: true})
		return nil
	case protocol.ProtocolOpenAIResponses:
		return output.Responses(run, resp.Body)
	case protocol.ProtocolOpenAIChatCompletions:
		return output.ChatCompletions(run, resp.Body)
	case protocol.ProtocolGeminiGenerateContent:
		return output.Gemini(run, resp.Body)
	default:
		return output.Error(run, "Unsupported native probe protocol")
	}
}

// payload builds protocol JSON using the prepared model and caller input.
// Structured decisions retain caller fields after model replacement and validation.
// Sampling attributes apply to text payloads using the same resolved route;
// unsupported protocol selections return an error before any transport call.
func (s *NativeProviderTest) payload(run *TestRun, value *provider.Record, probe nativeProbe, kind string) ([]byte, error) {
	model, prompt := probe.route.Model, probe.request.Prompt
	var object map[string]any
	switch probe.route.Protocol {
	case protocol.ProtocolGeminiGenerateContent:
		return geminiTestPayload(model, prompt, kind), nil
	case protocol.ProtocolSystemOne:
		return systemOneProbeBody(model, prompt, probe.request.SystemOne)
	case protocol.ProtocolAnthropicMessages:
		var err error
		object, err = anthropic.TestPayloadWithPrompt(model, prompt)
		if err != nil {
			return nil, err
		}
	case protocol.ProtocolOpenAIChatCompletions:
		object = openai.TestChatCompletionsPayload(model, prompt)
	case protocol.ProtocolOpenAIResponses:
		object = openai.TestResponsesPayload(model, prompt, false)
		// This probe supplies its own user prompt. Codex's internal system
		// instructions belong to its account endpoint and would alter this input.
		delete(object, "instructions")
	default:
		return nil, fmt.Errorf("unsupported native probe protocol %s", probe.route.Protocol)
	}
	if err := applyProbeSampling(run, s.Attributes, value, model, string(probe.route.Protocol), object); err != nil {
		return nil, err
	}
	return json.Marshal(object)
}

// nativeProbeEndpoint builds the selected protocol's streaming request URL.
// The caller supplies a validated root and final model. Gemini model actions
// validate that model using the shared AI Studio URL builder.
func nativeProbeEndpoint(probe nativeProbe) (string, error) {
	switch probe.route.Protocol {
	case protocol.ProtocolAnthropicMessages:
		return httpclient.BuildOpenAIEndpointURL(probe.baseURL, "/v1/messages"), nil
	case protocol.ProtocolOpenAIResponses:
		return httpclient.BuildOpenAIEndpointURL(probe.baseURL, "/v1/responses"), nil
	case protocol.ProtocolOpenAIChatCompletions:
		return httpclient.BuildOpenAIEndpointURL(probe.baseURL, "/v1/chat/completions"), nil
	case protocol.ProtocolGeminiGenerateContent:
		// Protocol roots include their version, while the shared AI Studio builder
		// accepts a service root and adds v1beta when constructing its model action.
		base := strings.TrimSuffix(strings.TrimRight(probe.baseURL, "/"), "/v1beta")
		return gemini.BuildGeminiAIStudioModelActionURL(base, probe.route.Model, "streamGenerateContent", true)
	default:
		return "", fmt.Errorf("unsupported streaming probe protocol %s", probe.route.Protocol)
	}
}

// send applies the provider's proxy and TLS selection to one prepared request.
// The caller owns the returned response body. Cancellation follows the request
// context and there is one transport attempt per invocation.
func (s *NativeProviderTest) send(req *http.Request, value *provider.Record) (*http.Response, error) {
	proxy := ""
	if value.ProxyID != nil && value.Proxy != nil {
		proxy = value.Proxy.URL()
	}
	var profile *tlsfingerprint.Profile
	if s.Profiles != nil {
		profile = s.Profiles.ResolveRequestTLS(egress.TLSSelection{Enabled: value.IsTLSFingerprintEnabled(), DirectProfileID: value.GetTLSFingerprintProfileID()})
	}
	return s.Transport.DoWithTLS(req, proxy, value.ID, value.Concurrency, profile)
}

// systemOneProbeBody builds a minimal decision or validates supplied structured input.
// Supplied state, questions, and extensions survive replacement of the model.
// Blank prompts use the protocol builder's availability question when structured
// input is absent. The caller supplies an already mapped model, which replaces
// any model in the supplied object before validation. This lets manual decision
// forms reuse their question body across several provider models while retaining
// the fields that describe each question and its evaluation state.
func systemOneProbeBody(model, prompt string, structured json.RawMessage) ([]byte, error) {
	if len(structured) == 0 {
		return systemone.ProbeBody(model, prompt), nil
	}
	// Supplied decisions can contain state and extensions unknown to the probe.
	// Replace just the model before protocol validation to retain those caller fields.
	body, err := systemone.ReplaceModel(structured, model)
	if err != nil {
		return nil, fmt.Errorf("invalid SystemOne test request: %w", err)
	}
	if _, err := systemone.ParseRequest(body); err != nil {
		return nil, err
	}
	return body, nil
}

// executeSystemOneProbe exchanges one prepared decision and emits its response.
// The target owns endpoint, authentication, and transport selection. The shared
// protocol executor closes its response and reports usage validity independently.
func executeSystemOneProbe(run *TestRun, target *jev.Target, body []byte) error {
	output := TestStreamOutput{}
	result, err := (jev.Executor{}).Execute(run.Context, upstream.AttemptInput{Protocol: protocol.ProtocolSystemOne, Body: body, Target: target}, nil)
	if err != nil {
		if run.Context.Err() != nil {
			return run.Context.Err()
		}
		return output.Error(run, "SystemOne test failed: "+err.Error())
	}
	output.SendEvent(run, provider.TestEvent{Type: "content", Text: string(result.MediaBody)})
	// Completion reports elapsed upstream execution in milliseconds alongside the
	// separately validated usage, so unknown billing evidence still permits visible answers.
	data := map[string]any{"duration_ms": result.Duration.Milliseconds(), "usage_valid": result.HasUsage, "response": json.RawMessage(result.MediaBody)}
	output.SendEvent(run, provider.TestEvent{Type: "test_complete", Model: result.UpstreamResponseModel, Success: true, Data: data})
	return nil
}
