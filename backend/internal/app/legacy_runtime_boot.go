package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/gateway"
	logger "github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"github.com/TokenFlux/TokenRouter/internal/modelcatalog/provider"
	"github.com/TokenFlux/TokenRouter/internal/server/runtimeconfig"
)

type bootRuntimeReady struct{}

func provideBootRuntime(
	catalog *provider.Service,
	manager *lifecycle.Manager,
	settingService *gateway.RuntimeSettings, forwarded *runtimeconfig.ForwardedSettings,
) *bootRuntimeReady {
	catalogReady := false
	manager.Register(lifecycle.Hook{Name: "ModelCatalogInitialization", StartOrder: 188, Start: func(context.Context) error {
		if catalog == nil {
			return nil
		}
		if err := catalog.Initialize(); err != nil {
			logger.LegacyPrintf("service.modelcatalog", "[Service] Warning: Model catalog initialization failed: %v", err)
			return nil
		}
		catalogReady = true
		return nil
	}})

	manager.Register(lifecycle.Hook{Name: "LegacySettingsInitialization", StartOrder: 181, StopOrder: 800, Start: func(ctx context.Context) error {
		if err := forwarded.LoadForwardedClientIPSettings(ctx); err != nil {
			logger.LegacyPrintf("service.setting", "Warning: load forwarded client IP settings failed: %v", err)
		}
		if err := settingService.MigrateGrokDefaultTextModel(ctx); err != nil {
			logger.LegacyPrintf("service.setting", "Warning: migrate Grok default text model failed: %v", err)
		}
		return nil
	}})

	manager.Register(lifecycle.Hook{Name: "ModelCatalogService", StartOrder: 980, StopOrder: 20, Start: func(ctx context.Context) error {
		if catalog != nil && catalogReady {
			catalog.Start()
		}
		return nil
	}, Stop: func(ctx context.Context) error {
		if catalog != nil {
			catalog.Stop()
		}
		return nil
	}})
	return &bootRuntimeReady{}
}
