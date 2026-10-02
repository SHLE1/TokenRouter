package httpapi

import (
	"context"
	"strings"
	"testing"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/scheduler/policy"
	"github.com/stretchr/testify/require"

	selectionadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider/selection"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// streamSelectionDiagnosticSource 为流执行后的下一次提供商选择提供调度查询数据。
// 诊断查询副本包含分组和活动状态，流夹具保存传输字段。
type streamSelectionDiagnosticSource struct {
	value gatewayprovider.ExecutionProvider
	group routing.Group
}

func (s streamSelectionDiagnosticSource) GetProvider(context.Context, int64) (*gatewayprovider.ExecutionProvider, error) {
	return &s.value, nil
}

func (s streamSelectionDiagnosticSource) GetGroup(context.Context, int64) (*routing.Group, error) {
	return &s.group, nil
}

func (s streamSelectionDiagnosticSource) ListProvidersForSchedulerScoreFilter(context.Context, string, string, string, string, int64, string) ([]gatewayprovider.ExecutionProvider, error) {
	return []gatewayprovider.ExecutionProvider{s.value}, nil
}

func (s streamSelectionDiagnosticSource) ListSchedulableProvidersForAdvancedSchedulerScore(context.Context, *int64, string) ([]gatewayprovider.ExecutionProvider, error) {
	return []gatewayprovider.ExecutionProvider{s.value}, nil
}

// selectionDiagnosticForStreamTest 通过实际诊断接口观察同一代理熔断实例。
func selectionDiagnosticForStreamTest(t *testing.T, source *OpenAIResponsesExecutor, value *gatewayprovider.ExecutionProvider) (bool, string) {
	t.Helper()
	copy := *value
	copy.Record.Status = "active"
	copy.Record.Schedulable = true
	copy.Record.GroupIDs = []int64{1}
	projection := streamSelectionDiagnosticSource{value: copy, group: routing.Group{ID: 1, Status: "active", Hydrated: true, SchedulerType: routing.GroupSchedulerTypeAdvanced}}
	parameters := scheduler.NewParameters(scheduler.NewSettingsRuntime(scheduler.Diagnostics{}), nil, scheduler.DefaultParameters())
	choices := selectionadapter.NewCompatible(selectionadapter.CompatibleDependencies{Responses: source.Lineage.Store, RuntimeBlocks: source.Output.Health.Runtime, ModelTransient: source.Output.Health.ModelTransient, ProxyCircuit: source.Output.ProxyCircuit}, selectionadapter.DefaultOptions())
	diagnostic := selectionadapter.NewDiagnostics(projection, selectionadapter.Shared{Parameters: parameters}, nil, choices)
	result, err := diagnostic.GetDetail(context.Background(), value.Record.ID, policy.AdvancedSchedulerScoreDiagnosticRequest{GroupID: 1})
	require.NoError(t, err)
	require.NotNil(t, result.Detail)
	return result.Detail.Eligible, strings.Join(result.Detail.HardFilterReasons, ",")
}
