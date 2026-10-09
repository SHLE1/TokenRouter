package app

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/TokenFlux/TokenRouter/internal/app/lifecycle"
	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"github.com/TokenFlux/TokenRouter/internal/requestlog"
	"github.com/TokenFlux/TokenRouter/internal/requestlog/postgres"
	"github.com/TokenFlux/TokenRouter/internal/requestlog/provider"
)

// provideRequestLog 为请求记录绑定持久队列、数据库和停机排空。
func provideRequestLog(db *sql.DB, cfg *config.Config, manager *lifecycle.Manager) (*requestlog.Service, error) {
	dir := cfg.RequestLog.SpoolDir
	if dir == "" {
		root := os.Getenv("DATA_DIR")
		if root == "" {
			root = cfg.Pricing.DataDir
		}
		dir = filepath.Join(root, "request-records")
	}
	queue, err := provider.NewQueue(dir)
	if err != nil {
		return nil, err
	}
	service := requestlog.NewService(postgres.NewStore(db), queue, cfg.RequestLog.RetentionDays,
		func(format string, args ...any) { logging.LegacyPrintf("request.records", format, args...) })
	manager.Register(lifecycle.Hook{Name: "RequestRecords", StartOrder: 185, StopOrder: 790, Start: service.Start, Stop: service.Stop})
	return service, nil
}
