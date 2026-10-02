package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"go.uber.org/zap"
)

// LogAPIKeyHealthWarning 将 API Key 健康事件写入应用日志。
func LogAPIKeyHealthWarning(message string, fields ...any) {
	attrs := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if ok {
			attrs = append(attrs, zap.Any(key, fields[i+1]))
		}
	}
	logging.L().Warn(message, attrs...)
}
