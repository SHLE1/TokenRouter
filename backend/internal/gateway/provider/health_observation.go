package provider

import (
	"context"
	"net/http"

	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// HealthObservationFromContext 读取模型、thinking 和端点信息，生成健康观测输入。
func HealthObservationFromContext(ctx context.Context, status int, headers http.Header, body []byte, models []string) provideradapter.HealthObservation {
	input := provideradapter.HealthObservation{Status: status, Headers: headers, Body: body, EffectiveModel: requeststate.HealthModel(ctx, models), ModelProvided: len(models) > 0, Thinking: requeststate.HealthThinking(ctx), ImagesEndpoint: requeststate.OpenAIImagesEndpointFromContext(ctx)}
	if len(models) > 0 {
		input.Model = models[0]
	}
	return input
}

// ApplyExecutionHealth 用独立记录执行健康观测，返回后将凭据和附加状态写回执行目标。
func ApplyExecutionHealth(ctx context.Context, observer *provideradapter.UpstreamHealth, target *ExecutionProvider, input provideradapter.HealthObservation) provider.UpstreamErrorDecision {
	record := ExecutionRecord(target)
	result := observer.ApplyUpstreamError(ctx, record, input)
	if target != nil && record != nil {
		target.Record.Credentials, target.Record.Extra = record.Credentials, record.Extra
	}
	return result
}

// ExecutionErrorPolicy 返回错误策略需要的提供商标识、平台、类型和凭据。
func ExecutionErrorPolicy(value *ExecutionProvider) *provider.Record {
	if value == nil {
		return nil
	}
	return &provider.Record{ID: value.Record.ID, Platform: value.Record.Platform, Type: value.Record.Type, Credentials: value.Record.Credentials}
}
