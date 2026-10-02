package app

import (
	"context"
	"database/sql"

	"github.com/TokenFlux/TokenRouter/internal/infra/postgres"
)

// databaseAdvisoryLease 绑定数据库连接来源，调用方决定锁身份和失败时的处理方式。
func databaseAdvisoryLease(db *sql.DB) func(context.Context, string) (func(), bool) {
	if db == nil {
		return nil
	}
	return func(ctx context.Context, key string) (func(), bool) {
		return postgres.TryAcquireDBAdvisoryLock(ctx, db, postgres.HashAdvisoryLockID(key))
	}
}
