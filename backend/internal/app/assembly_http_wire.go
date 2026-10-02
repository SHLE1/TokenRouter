//go:build wireinject

package app

import (
	serverhttp "github.com/TokenFlux/TokenRouter/internal/server/httpapi"

	"github.com/TokenFlux/TokenRouter/internal/server"

	"github.com/google/wire"
)

// httpAssemblyProviders 汇总HTTP 入口的 Wire provider。
var httpAssemblyProviders = wire.NewSet(
	nativeHTTPProviders,
	server.ProviderSet,
	provideHTTPOptions,
	provideHTTPRouteMount,
	provideAuthRouteMount,
	provideUserRouteMount,
	provideAdminRouteMount,
	provideGatewayRouteMount,
	providePaymentRouteMount,
	provideForwardedSettings,
	providePanelSettings,
	serverhttp.NewPanelSettingsHandler,
	provideRouterRuntime,
)
