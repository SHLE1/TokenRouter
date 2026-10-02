package postgres

import (
	"database/sql"
	"errors"

	dbent "github.com/TokenFlux/TokenRouter/ent"
	postgresinfra "github.com/TokenFlux/TokenRouter/internal/infra/postgres"
	infraerrors "github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
)

// translatePersistenceError 将未找到和唯一约束错误映射为调用方提供的领域错误。
// 对应的目标错误为 nil 或类型未匹配时返回输入错误，映射结果保留数据库错误原因。
func translatePersistenceError(err error, notFound, conflict *infraerrors.ApplicationError) error {
	if err == nil {
		return nil
	}

	// 兼容 Ent ORM 和标准 database/sql 的 NotFound 行为。
	// Ent 使用自定义的 NotFoundError，而标准库使用 sql.ErrNoRows。
	if notFound != nil && (errors.Is(err, sql.ErrNoRows) || dbent.IsNotFound(err)) {
		return notFound.WithCause(err)
	}

	// 处理唯一约束冲突（如邮箱已存在、名称重复等）
	if conflict != nil && postgresinfra.IsUniqueConstraintViolation(err) {
		return conflict.WithCause(err)
	}

	// 未匹配任何规则，返回原始错误
	return err
}
