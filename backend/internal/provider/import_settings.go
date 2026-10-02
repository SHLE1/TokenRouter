package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/provider/transfer"
)

// 导入模板保存允许的提供商默认配置。
type (
	OpenAIOAuthImportProviderDefaults = transfer.OpenAIOAuthImportProviderDefaults
	OpenAIOAuthImportDefaults         = transfer.OpenAIOAuthImportDefaults
)

// SettingKeyOpenAIOAuthImportDefaults 保留已有模板存储键。
const SettingKeyOpenAIOAuthImportDefaults = "openai_oauth_import_defaults"

// GetOpenAIOAuthImportDefaults 读取 OpenAI OAuth 导入模板。
func (s *RuntimeSettings) GetOpenAIOAuthImportDefaults(ctx context.Context) (*OpenAIOAuthImportDefaults, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyOpenAIOAuthImportDefaults)
	if err != nil {
		if errors.Is(err, s.notFound) {
			return DefaultOpenAIOAuthImportDefaults(), nil
		}
		return nil, fmt.Errorf("get openai oauth import defaults: %w", err)
	}
	if value == "" {
		return DefaultOpenAIOAuthImportDefaults(), nil
	}

	var settings OpenAIOAuthImportDefaults
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		slog.Warn("failed to unmarshal openai oauth import defaults, falling back to defaults",
			"error", err,
			"key", SettingKeyOpenAIOAuthImportDefaults)
		return DefaultOpenAIOAuthImportDefaults(), nil
	}

	return FillOpenAIOAuthImportDefaults(&settings), nil
}

// SetOpenAIOAuthImportDefaults 保存 OpenAI OAuth 导入模板。
func (s *RuntimeSettings) SetOpenAIOAuthImportDefaults(ctx context.Context, settings *OpenAIOAuthImportDefaults) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	if err := ValidateOpenAIOAuthImportDefaults(settings); err != nil {
		return err
	}
	// 模板副本与提供商写入使用相同的兼容清理规则。
	normalized := *settings
	normalized.Extra = maps.Clone(settings.Extra)
	NormalizeLegacyOpenAIProviderExtra(normalized.Extra)

	data, err := json.Marshal(&normalized)
	if err != nil {
		return fmt.Errorf("marshal openai oauth import defaults: %w", err)
	}

	return s.settingRepo.Set(ctx, SettingKeyOpenAIOAuthImportDefaults, string(data))
}

// FillOpenAIOAuthImportDefaults 补齐模板缺省字段，并规范化兼容配置。
func FillOpenAIOAuthImportDefaults(settings *OpenAIOAuthImportDefaults) *OpenAIOAuthImportDefaults {
	if settings == nil {
		return DefaultOpenAIOAuthImportDefaults()
	}

	defaults := DefaultOpenAIOAuthImportDefaults()
	// 读取模板时清理自动模式和探测字段，返回管理员保存的配置。
	settings.Extra = maps.Clone(settings.Extra)
	NormalizeLegacyOpenAIProviderExtra(settings.Extra)
	if len(defaults.Credentials) > 0 {
		if settings.Credentials == nil {
			settings.Credentials = map[string]any{}
		}
		for key, value := range defaults.Credentials {
			// 已保存的键保持当前值，空数组表示“不限制模型”。
			if _, exists := settings.Credentials[key]; !exists {
				settings.Credentials[key] = value
			}
		}
	}
	return settings
}

// FindForbiddenImportField 返回模板中禁止保存的字段名。
func FindForbiddenImportField(fields map[string]any, forbidden map[string]struct{}) (string, bool) {
	for key := range fields {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if _, ok := forbidden[normalized]; ok {
			return key, true
		}
	}
	return "", false
}

// ValidateOpenAIOAuthImportDefaults 校验导入默认模板。
func ValidateOpenAIOAuthImportDefaults(settings *OpenAIOAuthImportDefaults) error {
	if settings.Provider.Concurrency != nil && *settings.Provider.Concurrency < 0 {
		return fmt.Errorf("provider.concurrency must be >= 0")
	}
	if settings.Provider.Priority != nil && *settings.Provider.Priority < 0 {
		return fmt.Errorf("provider.priority must be >= 0")
	}
	if settings.Provider.RateMultiplier != nil && *settings.Provider.RateMultiplier < 0 {
		return fmt.Errorf("provider.rate_multiplier must be >= 0")
	}
	if settings.Provider.ExpiresAt != nil && *settings.Provider.ExpiresAt < 0 {
		return fmt.Errorf("provider.expires_at must be >= 0")
	}

	forbiddenCredentials := map[string]struct{}{
		"access_token":            {},
		"refresh_token":           {},
		"id_token":                {},
		"expires_at":              {},
		"email":                   {},
		"client_id":               {},
		"chatgpt_account_id":      {},
		"chatgpt_user_id":         {},
		"organization_id":         {},
		"plan_type":               {},
		"subscription_expires_at": {},
	}
	if field, ok := FindForbiddenImportField(settings.Credentials, forbiddenCredentials); ok {
		return fmt.Errorf("credentials.%s is not allowed in import defaults", field)
	}

	forbiddenExtra := map[string]struct{}{
		"email": {},
		"name":  {},
	}
	if field, ok := FindForbiddenImportField(settings.Extra, forbiddenExtra); ok {
		return fmt.Errorf("extra.%s is not allowed in import defaults", field)
	}

	return nil
}

// DefaultOpenAIOAuthImportDefaults 返回包含内置模型白名单的默认模板。
func DefaultOpenAIOAuthImportDefaults() *OpenAIOAuthImportDefaults {
	return &OpenAIOAuthImportDefaults{
		Credentials: map[string]any{
			"model_whitelist": []string{
				"gpt-5.2",
				"gpt-5.3",
				"gpt-5.3-spark",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-5.5",
			},
		},
	}
}
