package httpapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/stretchr/testify/require"
)

// TestTestEventSinkPreservesPreludeAndFrames 检查准备与提交阶段分别设置自适应及完整的 SSE Header。
func TestTestEventSinkPreservesPreludeAndFrames(t *testing.T) {
	recorder := httptest.NewRecorder()
	sink := NewTestEventSink(recorder)
	require.NoError(t, sink.Begin(context.Background(), false))
	require.False(t, recorder.Flushed)
	require.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
	require.Empty(t, recorder.Header().Get("Connection"))
	require.NoError(t, sink.Begin(context.Background(), true))
	require.True(t, recorder.Flushed)
	require.Equal(t, "keep-alive", recorder.Header().Get("Connection"))
	require.Equal(t, "no", recorder.Header().Get("X-Accel-Buffering"))
	require.NoError(t, sink.Emit(context.Background(), provider.TestEvent{Type: "content", Text: "line\n"}))
	require.Equal(t, "data: {\"type\":\"content\",\"text\":\"line\\n\"}\n\n", recorder.Body.String())
}
