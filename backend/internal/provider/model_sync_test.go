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

// TestModelSyncQueryPairsSuggestions installs a suggestion hook that reads the
// queried record and models. The result carries the hook's map beside an
// independently owned model slice, and a nil hook leaves suggestions absent.
func TestModelSyncQueryPairsSuggestions(t *testing.T) {
	models := []string{"first", "second"}
	svc := NewModelSyncService(func(context.Context, *Record) ([]string, error) { return models, nil })
	result, err := svc.Query(context.Background(), &Record{Platform: PlatformOpenCode})
	require.NoError(t, err)
	require.Nil(t, result.ModelProtocols)
	svc.Suggest = func(value *Record, ids []string) map[string]string {
		require.Equal(t, PlatformOpenCode, value.Platform)
		require.Equal(t, models, ids)
		return map[string]string{"first": "openai_responses"}
	}
	result, err = svc.Query(context.Background(), &Record{Platform: PlatformOpenCode})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"first": "openai_responses"}, result.ModelProtocols)
	result.Models[0] = "changed"
	require.Equal(t, "first", models[0])
}
