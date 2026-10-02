package composite

import "github.com/TokenFlux/TokenRouter/internal/moderation"

// ApplyModerationAdminReadSettings 将审核读取结果写入综合快照。
func (s *Snapshot) ApplyModerationAdminReadSettings(value *moderation.AdminReadSettings) {
	s.CyberSessionBlockEnabled = value.CyberSessionBlockEnabled
	s.CyberSessionBlockTTLSeconds = value.CyberSessionBlockTTLSeconds
	s.RiskControlEnabled = value.RiskControlEnabled
}
