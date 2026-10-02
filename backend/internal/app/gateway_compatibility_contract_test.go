package app

import (
	"context"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/stretchr/testify/require"
)

func TestSnapshotOpenAICompatibilityFallbackMetrics(t *testing.T) {
	before := gatewayCompatibilitySnapshot(nil)
	ctx := requeststate.WithThinkingEnabled(context.Background(), true)
	_, _ = requeststate.ThinkingEnabledFromContext(ctx)

	after := gatewayCompatibilitySnapshot(nil)
	// 请求使用提供商快照，响应返回约定的公开字段。
	require.Zero(t, after.MetadataTotal)
	require.Zero(t, before.MetadataTotal)
}
