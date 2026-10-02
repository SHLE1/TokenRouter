package composite

import "github.com/TokenFlux/TokenRouter/internal/provider"

// ApplyProviderAdminReadSettings 将提供商读取结果写入综合快照。
func (s *Snapshot) ApplyProviderAdminReadSettings(value *provider.AdminReadSettings) {
	s.ProviderQuotaNotifyEmails = value.ProviderQuotaNotifyEmails
	s.ProviderQuotaNotifyEnabled = value.ProviderQuotaNotifyEnabled
	s.ProviderSchedulingThresholds = value.ProviderSchedulingThresholds
}
