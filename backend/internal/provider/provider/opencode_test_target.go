package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/TokenFlux/TokenRouter/internal/upstream/opencode"
)

// opencode_test_target.go resolves and validates one OpenCode probe route before
// protocol execution. The loaded provider snapshot supplies both the one-hop
// model mapping and that final model's enabled protocol. Explicit selections
// and group routes must agree with the selected encoding, so validation and
// outbound model identity describe the same request.
//
// The target owns a single TestRun and OpenCode endpoint and authentication
// preparation. The protocol executor receives the resolved model directly and
// shares payload builders and stream readers with other probes. Group and
// client-entry identity remain available to attribute lookups, while each
// request's cancellation and sink failures belong to the original run.

// openCodeTestTarget retains shared probe dependencies and one provider configuration.
type openCodeTestTarget struct {
	executor *NativeProviderTest // shared protocol probe executor
	record   *provider.Record    // retained provider configuration
}

// Information returns OpenCode identity and its configured protocol selection.
func (t openCodeTestTarget) Information() provider.TestTargetInfo {
	return provider.TestTargetInfo{ProviderSnapshot: t.record.RoutingSnapshot()}
}

// Execute resolves one model/protocol pair and executes it through one TestRun.
// The loaded record stays stable throughout synchronous execution. Missing models,
// disabled protocols, and conflicting explicit selections emit an error before
// endpoint validation or credential transmission.
//
// Mapping reads the request model exactly once, including wildcard aliases. A
// mapped target such as b from a to b remains b even when b also maps to c.
// The selected group entry survives protocol conversion in the run's context,
// while the prepared native probe carries the final model and upstream protocol.
func (t openCodeTestTarget) Execute(ctx context.Context, request provider.PreparedTestRequest, sink provider.TestEventSink) error {
	headers := make(http.Header)
	headers.Set("User-Agent", request.UserAgent)
	run := NewTestRun(ctx, headers, sink)
	run.GroupID, run.Automatic = request.GroupID, request.Automatic
	defer run.Cancel()
	output := TestStreamOutput{}
	source := request.ProtocolRoute.ClientProtocol
	var fallbacks map[capability.ProtocolID][]capability.ProtocolID
	if source != "" {
		// Restrict conversion to the endpoint chosen during scheduling so refreshed
		// provider data cannot silently select a different protocol for this probe.
		fallbacks = map[capability.ProtocolID][]capability.ProtocolID{source: {request.ProtocolRoute.UpstreamProtocol}}
	}
	snapshot := t.record.RoutingSnapshot()
	resolved, ok := snapshot.ResolveModelRoute(strings.TrimSpace(request.Model), source, fallbacks)
	if !ok {
		return output.Error(run, fmt.Sprintf("OpenCode model %q has no enabled protocol; sync models or configure its protocol", request.Model))
	}
	if request.Protocol != "" && provider.TestProtocolID(request.Protocol) != resolved.Protocol {
		return output.Error(run, fmt.Sprintf("OpenCode model %q uses %s", resolved.Model, resolved.Protocol))
	}
	if request.ProtocolRoute.UpstreamProtocol != "" && request.ProtocolRoute.UpstreamProtocol != resolved.Protocol {
		return output.Error(run, "OpenCode model protocol changed")
	}
	if (request.TestType == provider.ProviderTestTypeDecision || len(request.SystemOne) > 0) && resolved.Protocol != capability.ProtocolSystemOne {
		return output.Error(run, "Decision tests require a SystemOne model")
	}
	if t.record.Type != provider.ProviderTypeAPIKey || strings.TrimSpace(t.record.GetCredential("api_key")) == "" {
		return output.Error(run, "OpenCode probes require an API key")
	}
	base, err := t.executor.ValidateURL((provider.ProtocolTarget{Record: t.record, Protocol: resolved.Protocol}).BaseURL())
	if err != nil {
		return output.Error(run, "Invalid OpenCode base URL: "+err.Error())
	}
	run.ProtocolRoute = request.ProtocolRoute
	if run.ProtocolRoute.ClientProtocol == "" {
		run.ProtocolRoute = provider.TestProtocolRoute{ClientProtocol: resolved.Protocol, UpstreamProtocol: resolved.Protocol}
	}
	run.Context = provider.WithTestProtocolRoute(run.Context, run.ProtocolRoute)
	probe := nativeProbe{
		route: resolved, request: request, baseURL: base,
		applyHeaders: func(headers http.Header) {
			ApplyProviderHeaderOverrides(t.record, headers)
			opencode.ApplyHeaders(headers, t.record.GetCredential("api_key"), run.GetHeader)
		},
	}
	return run.Result(t.executor.execute(run, t.record, probe))
}
