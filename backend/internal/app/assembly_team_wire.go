//go:build wireinject

package app

import (
	teamredis "github.com/TokenFlux/TokenRouter/internal/team/rediscache"

	"github.com/google/wire"
)

// teamAssemblyProviders 汇总 team 模块的 Wire provider。
var teamAssemblyProviders = wire.NewSet(
	teamHTTPProviders,
	provideTeam,
	provideTeamRepository,
	teamredis.NewTeamInvitationLimiter,
)
