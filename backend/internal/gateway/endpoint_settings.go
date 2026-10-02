package gateway

import "context"

// GetGrokDefaultBaseURLMode 在读取预算内查询 Grok Base URL 模式，缺省使用 CLI。
func (s *RuntimeSettings) GetGrokDefaultBaseURLMode(ctx context.Context) string {
	if s == nil || s.settingRepo == nil {
		return "cli"
	}
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewayForwardingDBTimeout)
	defer cancel()
	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyGrokDefaultBaseURLMode)
	if err != nil {
		return "cli"
	}
	return NormalizeGrokDefaultBaseURLMode(raw)
}
