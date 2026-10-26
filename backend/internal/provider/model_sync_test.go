package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// model_sync_test.go checks input isolation, catalogue ownership, and shutdown
// behavior. A live model adapter may mutate its input or delay returning after shutdown
// cancellation. A blocked fetch makes those effects observable, checking that
// the caller's credentials stay isolated and that stop rejects another query
// while retaining its first timeout result.

// TestModelSyncOwnsInflightQueryAndInput blocks a fetch after mutating its
// copied credential map, then times out shutdown. The original token stays
// intact, later queries fail as stopped, and the released query sees
// cancellation.
func TestModelSyncOwnsInflightQueryAndInput(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	completed := make(chan error, 1)
	core := NewModelSyncService(func(ctx context.Context, v *Record) ([]string, error) {
		v.Credentials["access_token"] = "changed-by-adapter"
		close(entered)
		<-release
		return nil, ctx.Err()
	})
	input := &Record{ID: 1, Credentials: map[string]any{"access_token": "original"}}
	// Keep the adapter in flight until stop has exhausted its budget. This
	// exposes both copied input ownership and the canceled context observed
	// when the deliberately blocked adapter finally returns.
	go func() {
		_, err := core.Query(context.Background(), input)
		completed <- err
	}()
	<-entered
	require.Equal(t, "original", input.Credentials["access_token"])
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	stopErr := core.StopContext(ctx)
	require.ErrorIs(t, stopErr, context.DeadlineExceeded)
	_, err := core.Query(context.Background(), input)
	require.ErrorIs(t, err, ErrRefreshStopped)
	close(release)
	require.ErrorIs(t, <-completed, context.Canceled)
	require.Equal(t, stopErr, core.StopContext(context.Background()))
}

// TestOpenCodeModelSyncQueryExcludesSavedOverrides queries a Go record with
// conflicting saved choices. The result contains current catalogue suggestions,
// leaves unknown models unresolved, and gives the caller independently owned data.
func TestOpenCodeModelSyncQueryExcludesSavedOverrides(t *testing.T) {
	models := []string{"new", "custom", "unknown", "minimax-m3"}
	routes := map[string]string{"new": "openai_responses", "custom": "openai_chat_completions"}
	svc := NewModelSyncService(func(context.Context, *Record) ([]string, error) { return models, nil })
	svc.CatalogProtocols = func(mode string) map[string]string {
		require.Equal(t, "go", mode)
		return routes
	}
	value := &Record{
		Platform: PlatformOpenCode, Type: ProviderTypeAPIKey,
		Credentials: map[string]any{
			"provider_mode": "go",
			OpenCodeModelProtocolsKey: map[string]any{
				"go":   map[string]any{"custom": "anthropic_messages"},
				"payg": map[string]any{"new": "systemone"},
			},
			"opencode_catalog_protocols": map[string]any{"go": map[string]any{"unknown": "systemone"}},
		},
	}
	result, err := svc.Query(context.Background(), value)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"new": "openai_responses", "custom": "openai_chat_completions"}, result.ModelProtocols)
	result.Models[0] = "changed"
	result.ModelProtocols["new"] = "changed"
	require.Equal(t, "new", models[0])
	require.Equal(t, "openai_responses", routes["new"])
}

// TestOpenCodeModelSyncUsesBundledSuggestions queries without a live publication.
// Bundled Go routes remain available, and saved manual values cannot replace them.
func TestOpenCodeModelSyncUsesBundledSuggestions(t *testing.T) {
	svc := NewModelSyncService(func(context.Context, *Record) ([]string, error) { return []string{"minimax-m3"}, nil })
	value := &Record{
		Platform: PlatformOpenCode, Type: ProviderTypeAPIKey,
		Credentials: map[string]any{
			"provider_mode":           "go",
			OpenCodeModelProtocolsKey: map[string]any{"go": map[string]any{"minimax-m3": "systemone"}},
		},
	}
	result, err := svc.Query(context.Background(), value)
	require.NoError(t, err)
	require.Equal(t, "anthropic_messages", result.ModelProtocols["minimax-m3"])
}

// TestModelSyncPreviewOwnsTemporaryCredentials supplies a temporary root and key.
// The use case builds the complete unsaved query record and returns paired model results.
func TestModelSyncPreviewOwnsTemporaryCredentials(t *testing.T) {
	svc := NewModelSyncService(func(_ context.Context, value *Record) ([]string, error) {
		require.Zero(t, value.ID)
		require.Equal(t, "temporary-key", value.GetCredential("api_key"))
		require.Equal(t, "https://preview.example", value.OpenCodeRoot())
		require.Equal(t, "go", value.OpenCodeMode())
		return []string{"minimax-m3"}, nil
	})
	result, err := svc.Preview(context.Background(), ModelSyncPreview{
		Platform: PlatformOpenCode, Type: ProviderTypeAPIKey, APIKey: "temporary-key",
		BaseURL: "https://preview.example", ProviderMode: "go",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"minimax-m3"}, result.Models)
	require.Equal(t, "anthropic_messages", result.ModelProtocols["minimax-m3"])
}
