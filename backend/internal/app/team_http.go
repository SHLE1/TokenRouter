//go:build wireinject

package app

import (
	teamhttp "github.com/TokenFlux/TokenRouter/internal/team/httpapi"
	"github.com/google/wire"
)

// teamHTTPProviders 汇总团队用户端和管理端 HTTP 构造函数。
var teamHTTPProviders = wire.NewSet(teamhttp.NewUserHandler, teamhttp.NewAdminHandler)
