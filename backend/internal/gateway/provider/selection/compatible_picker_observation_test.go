package selection

import (
	"context"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	schedulercore "github.com/TokenFlux/TokenRouter/internal/scheduler"
)

func (s *compatiblePicker) isProviderRequestCompatible(ctx context.Context, provider *gatewayprovider.ExecutionProvider, req schedulercore.PlatformSelectionInput) bool {
	compatible, _ := s.isProviderRequestCompatibleReason(ctx, provider, req)
	return compatible
}
