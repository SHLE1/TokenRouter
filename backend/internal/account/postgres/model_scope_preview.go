package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/TokenFlux/TokenRouter/internal/account"
)

// ModelScopeMigrationWarning 区分显式目录扩展和升级后不再接受的近期模型，不返回凭据。
type ModelScopeMigrationWarning struct {
	AccountID                 int64    `json:"account_id"`
	ConfiguredOutsideDefaults []string `json:"configured_outside_defaults,omitempty"`
	Models                    []string `json:"recent_models_outside_scope,omitempty"`
}

// PreviewModelScopeMigration 核对全部账号的显式模型和最近三十天使用记录，预检不访问上游。
func PreviewModelScopeMigration(ctx context.Context, db *sql.DB, supports func(*account.Record, string) bool, configuredOutsideDefaults func(*account.Record) []string) ([]ModelScopeMigrationWarning, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `
WITH recent_models AS (
 SELECT account_id, jsonb_agg(DISTINCT COALESCE(NULLIF(upstream_model,''),model) ORDER BY COALESCE(NULLIF(upstream_model,''),model)) AS models
 FROM usage_logs WHERE created_at >= NOW()-INTERVAL '30 days' GROUP BY account_id
)
SELECT a.id,a.platform,a.type,a.credentials,a.extra,a.parent_account_id,COALESCE(m.models,'[]'::jsonb)
FROM accounts a LEFT JOIN recent_models m ON m.account_id=a.id
WHERE a.deleted_at IS NULL ORDER BY a.id`)
	if err != nil {
		return nil, err
	}
	warnings := make([]ModelScopeMigrationWarning, 0)
	for rows.Next() {
		var value account.Record
		var credentials, extra, rawModels []byte
		if err := rows.Scan(&value.ID, &value.Platform, &value.Type, &credentials, &extra, &value.ParentAccountID, &rawModels); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if err := json.Unmarshal(credentials, &value.Credentials); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("decode model policy for account %d: %w", value.ID, err)
		}
		if len(extra) > 0 {
			if err := json.Unmarshal(extra, &value.Extra); err != nil {
				_ = rows.Close()
				return nil, fmt.Errorf("decode account %d options: %w", value.ID, err)
			}
		}
		warning := ModelScopeMigrationWarning{AccountID: value.ID, ConfiguredOutsideDefaults: configuredOutsideDefaults(&value)}
		// 已明确全模型透传的历史账号会被迁移写成显式通配范围。
		if value.IsOpenAIPassthroughEnabled() {
			if whitelist, explicit := account.ResolveFinalModelWhitelist(value.Platform, value.Credentials, nil); !explicit || len(whitelist) == 0 {
				if value.Credentials == nil {
					value.Credentials = map[string]any{}
				}
				value.Credentials["model_whitelist"] = []string{"*"}
			}
		}
		var models []string
		if err := json.Unmarshal(rawModels, &models); err != nil {
			_ = rows.Close()
			return nil, err
		}
		for _, model := range models {
			if !supports(&value, model) {
				warning.Models = append(warning.Models, model)
			}
		}
		if len(warning.ConfiguredOutsideDefaults) > 0 || len(warning.Models) > 0 {
			warnings = append(warnings, warning)
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	return warnings, tx.Commit()
}
