package app

import (
	"context"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// newOpenAIExecutionCredentialsForTest 使用夹具传入的存储和 token 源。
func newOpenAIExecutionCredentialsForTest(repo gatewayprovider.ExecutionProviderStore, grok *provider.GrokTokenSource) *provider.OpenAIExecutionCredentials {
	out := &provider.OpenAIExecutionCredentials{}
	if repo != nil {
		out.Parent = func(ctx context.Context, id int64) (*provider.Record, error) {
			value, err := repo.GetByID(ctx, id)
			return gatewayprovider.ExecutionRecord(value), err
		}
	}
	if grok != nil {
		out.Grok = grok.GetAccessToken
	}
	return out
}
