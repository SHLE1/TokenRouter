package composite

import "github.com/TokenFlux/TokenRouter/internal/creative"

// ApplyCreativeAdminReadSettings 将创作台读取结果写入综合快照。
func (s *Snapshot) ApplyCreativeAdminReadSettings(value *creative.AdminReadSettings) {
	s.CreativeEnabled = value.CreativeEnabled
	s.CreativeModelSettings = value.CreativeModelSettings
	s.CreativeWorkerCount = value.CreativeWorkerCount
}
