package app

import (
	"log/slog"

	"github.com/TokenFlux/TokenRouter/internal/gateway/promptpolicy"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

// provideGatewayPromptPolicy 绑定设置表和诊断日志，提示词策略由服务缓存并更新。
func provideGatewayPromptPolicy(store *settings.Store) *promptpolicy.Service {
	return promptpolicy.New(store, settings.ErrSettingNotFound, slog.Warn)
}
