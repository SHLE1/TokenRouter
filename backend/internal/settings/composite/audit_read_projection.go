package composite

import "github.com/TokenFlux/TokenRouter/internal/audit"

// ApplyAuditAdminReadSettings 将审计读取结果写入综合快照。
func (s *Snapshot) ApplyAuditAdminReadSettings(value *audit.AdminReadSettings) {
	s.AuditLogRetentionDays = value.AuditLogRetentionDays
}
