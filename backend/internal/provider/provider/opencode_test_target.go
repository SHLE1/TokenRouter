package provider

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// opencode_test_target.go selects one model's native probe executor for an
// OpenCode provider. The final mapped model determines the protocol for the
// selected service mode, and both explicit protocol requests and routed group
// probes must agree with that encoding before an upstream call is made.
//
// Gemini and SystemOne reuse their dedicated probe implementations. The three
// other encodings use the existing native API-key probe executor with a single
// selected protocol, so an automatic test cannot accidentally send one model
// to every protocol the provider supports. The retained record and executors
// are read-only during each prepared probe.

// openCodeTestTarget retains shared probe executors and one provider configuration.
type openCodeTestTarget struct {
	targets *TestTargets     // shared native probe executors
	record  *provider.Record // retained provider configuration
}

// Information returns OpenCode identity and its configured protocol selection.
func (t openCodeTestTarget) Information() provider.TestTargetInfo {
	return provider.TestTargetInfo{ProviderSnapshot: t.record.RoutingSnapshot()}
}

// Execute resolves the final model and delegates one native protocol probe.
// Missing models and conflicting explicit or routed protocol selections fail
// before sending credentials. Validation failures emit a test error before
// returning to the shared service.
func (t openCodeTestTarget) Execute(ctx context.Context, request provider.PreparedTestRequest, sink provider.TestEventSink) error {
	headers := make(http.Header)
	headers.Set("User-Agent", request.UserAgent)
	run := NewTestRun(ctx, headers, sink)
	defer run.Cancel()
	model := mappedTestModel(t.record, request.Model)
	target, known := t.record.OpenCodeModelProtocols()[model]
	if !known || !slices.Contains(t.record.UpstreamProtocols(), target) {
		return (TestStreamOutput{}).Error(run, fmt.Sprintf("OpenCode model %q has no enabled protocol; sync models or configure its protocol", model))
	}
	if request.Protocol != "" && provider.TestProtocolID(request.Protocol) != target {
		return (TestStreamOutput{}).Error(run, fmt.Sprintf("OpenCode model %q uses %s", model, target))
	}
	if request.ProtocolRoute.UpstreamProtocol != "" && request.ProtocolRoute.UpstreamProtocol != target {
		return (TestStreamOutput{}).Error(run, "OpenCode model protocol changed")
	}
	if (request.TestType == provider.ProviderTestTypeDecision || len(request.SystemOne) > 0) && target != capability.ProtocolSystemOne {
		return (TestStreamOutput{}).Error(run, "Decision tests require a SystemOne model")
	}
	switch target {
	case capability.ProtocolGeminiGenerateContent:
		return t.targets.Gemini.Target(t.record).Execute(ctx, request, sink)
	case capability.ProtocolSystemOne:
		return t.targets.Jev.Target(t.record).Execute(ctx, request, sink)
	default:
		run.GroupID, run.Automatic = request.GroupID, request.Automatic
		run.ProtocolRoute = request.ProtocolRoute
		if run.ProtocolRoute.ClientProtocol == "" {
			// Standalone probes use their native encoding as the client entry;
			// routed group probes retain their original entry for scoped attribute rules.
			run.ProtocolRoute = provider.TestProtocolRoute{ClientProtocol: target, UpstreamProtocol: target}
		}
		return run.Result(t.targets.CN.ExecuteAdaptive(run, t.record, request.Model, request.Prompt, target))
	}
}
