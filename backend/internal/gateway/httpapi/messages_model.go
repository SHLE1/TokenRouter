package httpapi

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway/modeltrace"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

// ResolveOpenAIMessagesAccountLayerModel 将通用分组映射结果规范化后交给账号模型规则。
func ResolveOpenAIMessagesAccountLayerModel(groupMappedModel string) string {
	return provider.NormalizeOpenAICompatRequestedModel(groupMappedModel)
}

// ResolveOpenAIMessagesAccountLayerModelForRequest 登记规范化结果，供响应恢复和用量追踪使用。
func ResolveOpenAIMessagesAccountLayerModelForRequest(ctx context.Context, groupMappedModel string) string {
	model := ResolveOpenAIMessagesAccountLayerModel(groupMappedModel)
	modeltrace.RegisterStage(ctx, model)
	return model
}
