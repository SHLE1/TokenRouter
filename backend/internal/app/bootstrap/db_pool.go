package bootstrap

import (
	"database/sql"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/infra/postgres"
)

func postgresPoolOptions(cfg *config.Config) postgres.PoolOptions {
	return postgres.PoolOptions{
		MaxOpenConns:           cfg.Database.MaxOpenConns,
		MaxIdleConns:           cfg.Database.MaxIdleConns,
		ConnMaxLifetimeMinutes: cfg.Database.ConnMaxLifetimeMinutes,
		ConnMaxIdleTimeMinutes: cfg.Database.ConnMaxIdleTimeMinutes,
	}
}

func applyDBPoolSettings(db *sql.DB, cfg *config.Config) {
	postgres.ApplyPoolSettings(db, postgresPoolOptions(cfg))
}
