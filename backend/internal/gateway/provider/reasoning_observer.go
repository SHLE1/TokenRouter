package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"go.uber.org/zap"
)

// WarnReasoningCacheFailure 记录推理缓存失败的诊断字段。
func WarnReasoningCacheFailure(itemID string, err error) {
	logging.L().Warn("openai responses chat fallback: cache reasoning content failed", zap.Error(err), zap.String("item_id", itemID))
}
