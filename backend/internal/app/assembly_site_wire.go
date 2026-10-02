//go:build wireinject

package app

import (
	"github.com/TokenFlux/TokenRouter/internal/site"

	sitepostgres "github.com/TokenFlux/TokenRouter/internal/site/postgres"
	"github.com/google/wire"
)

// siteAssemblyProviders 汇总 site 模块的 Wire provider。
var siteAssemblyProviders = wire.NewSet(
	provideSiteDisplay,
	provideSitePublicHTTP,
	provideSitePublic,
	provideSitePages,
	site.NewAnnouncementService,
	sitepostgres.NewAnnouncementRepository,
	sitepostgres.NewAnnouncementReadRepository,
	provideAnnouncementUsers,
	provideAnnouncementSubscriptions,
	provideAnnouncementExpiry,
)
