package audit

// AdminReadSettings 包含审计日志保留天数。
type AdminReadSettings struct{ AuditLogRetentionDays int }

// ReadAdminSettings 从传入的设置值解析审计日志保留天数。
func ReadAdminSettings(settings map[string]string) *AdminReadSettings {
	result := &AdminReadSettings{}
	result.AuditLogRetentionDays = ParseRetentionDays(settings[SettingKeyAuditLogRetentionDays])

	return result
}
