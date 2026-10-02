package provider

// SanitizeStoredCredentials 移除 OAuth 兑换后不得写入提供商凭据的 Grok SSO、密码和 cookie。
// 批量操作可能缺少平台标识，所有平台都清理 cookie。
func SanitizeStoredCredentials(platform string, creds map[string]any) map[string]any {
	if creds == nil {
		return nil
	}
	_ = platform
	for _, key := range []string{
		"password", "sso_token", "sso", "sso-rw", "clearTextPassword", "cookie",
	} {
		delete(creds, key)
	}
	return creds
}
