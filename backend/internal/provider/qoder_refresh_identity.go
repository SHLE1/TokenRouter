package provider

// QoderRefreshCredentialsHash 计算刷新身份字段的哈希，供凭据竞争检查使用。
func QoderRefreshCredentialsHash(credentials map[string]any) string {
	if len(credentials) == 0 {
		return QoderCredentialsHash(nil)
	}
	keys := []string{
		"pat",
		"security_oauth_token",
		"refresh_token",
		"machine_id",
		"machine_token",
		"machine_type",
		"uid",
		"aid",
		"organization_id",
		"organization_name",
		"name",
		"user_type",
		"site",
		"refresh_mode",
	}
	auth := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := credentials[key]; ok {
			auth[key] = value
		}
	}
	return QoderCredentialsHash(auth)
}
