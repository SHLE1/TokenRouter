package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// opencode_test_target_test.go verifies that OpenCode probe preparation rejects
// unknown models and conflicting operations before selecting an executor. The
// fixtures leave every executor unavailable, making an accidental upstream
// dispatch fail instead of silently accepting the invalid request.

// openCodeProbeSink records preparation failures before any native executor runs.
type openCodeProbeSink struct {
	events []provider.TestEvent // emitted probe observations
}

// Begin accepts probe preparation for this in-memory event recorder.
func (*openCodeProbeSink) Begin(context.Context, bool) error { return nil }

// Emit retains one probe event for subsequent sequential assertions.
func (s *openCodeProbeSink) Emit(_ context.Context, event provider.TestEvent) error {
	s.events = append(s.events, event)
	return nil
}

// TestOpenCodeProbeRejectsUnknownAndConflictingRoutes exercises preparation with
// unavailable native executors. Every invalid request produces an error event
// before any executor is invoked or credentials could be sent upstream.
func TestOpenCodeProbeRejectsUnknownAndConflictingRoutes(t *testing.T) {
	record := &provider.Record{Platform: provider.PlatformOpenCode, Type: provider.ProviderTypeAPIKey, Credentials: map[string]any{"api_key": "test", "opencode_model_protocols": map[string]any{"payg": map[string]any{"custom": "openai_chat_completions"}}}}
	target := openCodeTestTarget{record: record}
	for _, request := range []provider.PreparedTestRequest{
		{TestRequest: provider.TestRequest{Model: "unknown"}},
		{TestRequest: provider.TestRequest{Model: "custom", Protocol: "anthropic"}},
		{TestRequest: provider.TestRequest{Model: "custom"}, TestType: provider.ProviderTestTypeDecision},
	} {
		sink := &openCodeProbeSink{}
		require.Error(t, target.Execute(t.Context(), request, sink))
		require.Len(t, sink.events, 1)
		require.Equal(t, "error", sink.events[0].Type)
	}
}
