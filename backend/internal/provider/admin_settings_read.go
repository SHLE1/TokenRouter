package provider

import (
	"log/slog"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/identity/contact"
)

// AdminReadSettings 保存综合管理页展示的提供商设置。
type AdminReadSettings struct {
	ProviderQuotaNotifyEmails    []contact.Entry
	ProviderQuotaNotifyEnabled   bool
	ProviderSchedulingThresholds map[string]int
}

// ReadAdminSettings 从同一批持久化数据解析展示值，并处理字段缺省值。
func ReadAdminSettings(settings map[string]string) *AdminReadSettings {
	result := &AdminReadSettings{}

	result.ProviderQuotaNotifyEnabled = settings[SettingKeyProviderQuotaNotifyEnabled] == "true"
	if raw := strings.TrimSpace(settings[SettingKeyProviderQuotaNotifyEmails]); raw != "" {
		result.ProviderQuotaNotifyEmails = contact.ParseNotifyEmails(raw)
	}
	if result.ProviderQuotaNotifyEmails == nil {
		result.ProviderQuotaNotifyEmails = []contact.Entry{}
	}
	result.ProviderSchedulingThresholds = DefaultProviderSchedulingThresholds()
	if raw := strings.TrimSpace(settings[SettingKeyProviderSchedulingThresholds]); raw != "" {
		if thresholds, err := ParseProviderSchedulingThresholdsSetting(raw); err != nil {
			slog.Warn("[Setting] parseSettings: unmarshal provider_scheduling_thresholds failed", "error", err)
		} else {
			result.ProviderSchedulingThresholds = thresholds
		}
	}
	return result
}
