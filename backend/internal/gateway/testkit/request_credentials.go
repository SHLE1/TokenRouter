package testkit

import (
	"context"
	"log/slog"
	"time"

	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// RequestCredentials 组合测试提供的凭据读取和刷新接口。
func RequestCredentials(store gatewayadapter.ExecutionProviderStore, source *provider.OpenAIExecutionCredentials, tokens *provider.GrokTokenSource, blocks *provider.RuntimeBlockState) *gatewayadapter.RequestCredentials {
	if blocks == nil {
		blocks = provider.NewRuntimeBlockState(time.Now)
	}
	if source == nil {
		source = &provider.OpenAIExecutionCredentials{}
	}
	recovery := &provider.GrokCredentialRecovery{Runtime: blocks, Warn: slog.Warn}
	if store != nil {
		source.Parent = func(ctx context.Context, id int64) (*provider.Record, error) {
			value, err := store.GetByID(ctx, id)
			return gatewayadapter.ExecutionRecord(value), err
		}
		recovery.Read = source.Parent
		recovery.State, _ = store.(provider.GrokCredentialStateWriter)
	}
	if tokens != nil {
		source.Grok = tokens.GetAccessToken
		recovery.Invalidate = tokens.InvalidateToken
	}
	return &gatewayadapter.RequestCredentials{Source: source, HasGrokTokenSource: tokens != nil, Recovery: recovery, Runtime: blocks}
}
