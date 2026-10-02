package provider

// sparkShadowAllowedCredentialKeys 是 spark 影子提供商唯一可写的凭据键集合(仅模型映射)。
// 凭据校验与清理共用此列表。
var sparkShadowAllowedCredentialKeys = map[string]struct{}{
	"model_mapping":         {},
	"compact_model_mapping": {},
}

func IsAllowedSparkShadowCredentialsUpdate(credentials map[string]any) bool {
	if credentials == nil {
		return true
	}
	for key := range credentials {
		if _, ok := sparkShadowAllowedCredentialKeys[key]; !ok {
			return false
		}
	}
	return true
}

func SanitizeSparkShadowCredentials(credentials map[string]any) map[string]any {
	if len(credentials) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(sparkShadowAllowedCredentialKeys))
	for key := range sparkShadowAllowedCredentialKeys {
		if value, ok := credentials[key]; ok && value != nil {
			out[key] = value
		}
	}
	return out
}
