package modeltrace

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// WithGroupRedirect 组合 Key 和分组的模型映射，记录本次请求恢复响应模型时所需的名称。
func WithGroupRedirect(result routing.GroupMappingResult, ctx context.Context, requestedModel string) routing.GroupMappingResult {
	trace, ok := FromContext(ctx)
	if !ok {
		return result
	}
	result.ClientModel = trace.ClientModel
	result.APIKeyRedirected = true
	RegisterStage(ctx, requestedModel)
	RegisterStage(ctx, result.MappedModel)
	return result
}

// RegisterStage 登记当前请求已使用的内部模型，供响应元数据恢复。
func RegisterStage(ctx context.Context, model string) {
	if trace, ok := FromContext(ctx); ok {
		trace.RegisterModel(model)
	}
}
