package provider

import (
	"context"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/stretchr/testify/require"
)

// TestRoutePlannerWithoutGroupDoesNotReadConfiguration 检查未绑定分组时是否返回请求模型，存储读取次数为零。
func TestRoutePlannerWithoutGroupDoesNotReadConfiguration(t *testing.T) {
	planner := NewRoutePlanner(routing.NewPricingConfigService(nil, nil))
	plan := planner.PlanRoute(context.Background(), nil, nil, "request-model")
	require.Equal(t, routing.GroupMappingResult{MappedModel: "request-model"}, plan.Mapping())
	require.Zero(t, plan.GroupID())
}
