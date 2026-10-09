package app

import (
	"database/sql"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"github.com/TokenFlux/TokenRouter/internal/requestlog"
	"github.com/TokenFlux/TokenRouter/internal/requestlog/postgres"
)

// provideRequestLog 为请求记录绑定数据库和后台批写的启停。
func provideRequestLog(db *sql.DB, cfg *config.Config, manager *lifecycle.Manager) *requestlog.Service {
	service := requestlog.NewService(postgres.NewStore(db), cfg.RequestLog.RetentionDays,
		func(format string, args ...any) { logging.LegacyPrintf("request.records", format, args...) })
	manager.Register(lifecycle.Hook{Name: "RequestRecords", StartOrder: 185, StopOrder: 790, Start: service.Start, Stop: service.Stop})
	return service
}
