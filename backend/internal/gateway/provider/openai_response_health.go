package provider

import (
	"context"
	"net/http"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// ApplyOpenAIResponseHealth 汇总请求类型、模型和响应信息，交给提供商健康策略处理。
func ApplyOpenAIResponseHealth(ctx context.Context, health *provideradapter.OpenAIResponseHealth, target *ExecutionProvider, status int, headers http.Header, body []byte, suppressDefaultRateLimit bool, models ...string) provider.UpstreamErrorDecision {
	return health.Apply(ctx, target.View(), provideradapter.OpenAIResponseHealthInput{
		Observation:              HealthObservationFromContext(ctx, status, headers, body, models),
		Models:                   models,
		SuppressDefaultRateLimit: suppressDefaultRateLimit,
		ContentRejected:          IsContentPolicyRejection(body) || IsOpenAICyberWarningPayload(body, upstream.ExtractErrorMessage(body)),
		RequestScoped:            IsRequestScopedProviderFailure(target, status, body),
		SelfBuiltImage:           openai.IsOpenAIImagesSelfBuiltRequest(ctx),
		Transient:                IsTransientProviderFailure(status, body),
	})
}
