package bootstrap

import (
	"context"
	"database/sql"

	postgresinfra "github.com/TokenFlux/TokenRouter/internal/infra/postgres"
	"github.com/TokenFlux/TokenRouter/migrations"
)

// ApplyMigrations 执行数据库迁移。
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	return postgresinfra.ApplyMigrations(ctx, db, migrations.FS)
}
