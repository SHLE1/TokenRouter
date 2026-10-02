package selection

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// isProviderRequestCompatible 将兼容性判断结果转换为测试断言使用的布尔值。
func (s *compatiblePicker) isProviderRequestCompatible(ctx context.Context, provider *provider.ExecutionProvider, req scheduler.PlatformSelectionInput) bool {
	compatible, _ := s.isProviderRequestCompatibleReason(ctx, provider, req)
	return compatible
}
