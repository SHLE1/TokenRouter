package bootstrap

import (
	"context"
	"database/sql"

	"github.com/TokenFlux/TokenRouter/internal/identity"
	ip "github.com/TokenFlux/TokenRouter/internal/identity/postgres"
)

// CreateInitialAdmin 装配安装所需的身份存储并创建管理员。
func CreateInitialAdmin(ctx context.Context, db *sql.DB, input identity.InitialAdminInput, password func() (string, error)) (bool, string, error) {
	return ip.CreateInitialAdmin(ctx, db, input, password)
}
