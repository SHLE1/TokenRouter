package httpapi

import (
	"context"
	"net/http"

	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// ApplyHTTPFailure 使用传入的首个模型处理 HTTP 失败。
func (p *OpenAIResponseOutput) ApplyHTTPFailure(ctx context.Context, resp *http.Response, target *gatewayadapter.ExecutionProvider, body []byte, models ...string) provider.UpstreamErrorDecision {
	if len(models) > 0 {
		return gatewayadapter.ApplyOpenAIResponseHealth(ctx, p.Health, target, resp.StatusCode, resp.Header, body, false, models[0])
	}
	return gatewayadapter.ApplyOpenAIResponseHealth(ctx, p.Health, target, resp.StatusCode, resp.Header, body, false)
}
