package gateway

import "context"

// IsIdentityPatchEnabled 在请求时读取身份修补开关，读取失败时返回 true。
func (s *RuntimeSettings) IsIdentityPatchEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyEnableIdentityPatch)
	if err != nil {
		return true
	}
	return value == "true"
}

// GetIdentityPatchPrompt 保留读取失败或空值时使用平台默认提示词的约定。
func (s *RuntimeSettings) GetIdentityPatchPrompt(ctx context.Context) string {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyIdentityPatchPrompt)
	if err != nil {
		return ""
	}
	return value
}
