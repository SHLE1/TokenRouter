package composite

import "github.com/TokenFlux/TokenRouter/internal/notification"

// ApplyNotificationAdminReadSettings 将邮件读取结果写入综合快照。
func (s *Snapshot) ApplyNotificationAdminReadSettings(value *notification.AdminReadSettings) {
	s.SMTPFrom = value.SMTPFrom
	s.SMTPFromName = value.SMTPFromName
	s.SMTPHost = value.SMTPHost
	s.SMTPPassword = value.SMTPPassword
	s.SMTPPasswordConfigured = value.SMTPPasswordConfigured
	s.SMTPPort = value.SMTPPort
	s.SMTPUseTLS = value.SMTPUseTLS
	s.SMTPUsername = value.SMTPUsername
}
