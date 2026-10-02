package selection

import (
	"context"
	"maps"
	"strings"

	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	schedulercore "github.com/TokenFlux/TokenRouter/internal/scheduler"
	"github.com/TokenFlux/TokenRouter/internal/scheduler/policy"
)

// pickerEngine 包含候选评分、计数和反馈操作。
type pickerEngine interface {
	Select(context.Context, schedulercore.PlatformSelectionInput) (*gatewayadapter.SelectionResult, schedulercore.PlatformDecision, error)
	ReportResult(int64, bool, *int, ...policy.FeedbackConfig)
	ReportSwitch()
	SnapshotMetrics() schedulercore.PlatformMetricsSnapshot
}

// requestRoutingModel 返回提供商层模型，未提供时使用客户端模型。
func requestRoutingModel(input schedulercore.PlatformSelectionInput) string {
	if model := strings.TrimSpace(input.RoutingModel); model != "" {
		return model
	}
	return input.RequestedModel
}

// cloneSelectionInput 复制本次选择输入和排除集合。
func cloneSelectionInput(input schedulercore.PlatformSelectionInput) schedulercore.PlatformSelectionInput {
	input.ExcludedIDs = maps.Clone(input.ExcludedIDs)
	return input
}
