package composite

import "github.com/TokenFlux/TokenRouter/internal/promotion"

// ApplyPromotionAdminReadSettings 将推广读取结果写入综合快照。
func (s *Snapshot) ApplyPromotionAdminReadSettings(value *promotion.AdminReadSettings) {
	s.AdminRechargeRebateEnabled = value.AdminRechargeRebateEnabled
	s.AffiliateEnabled = value.AffiliateEnabled
	s.AffiliateRebateDurationDays = value.AffiliateRebateDurationDays
	s.AffiliateRebateFreezeHours = value.AffiliateRebateFreezeHours
	s.AffiliateRebatePerInviteeCap = value.AffiliateRebatePerInviteeCap
	s.AffiliateRebateRate = value.AffiliateRebateRate
	s.InvitationCodeEnabled = value.InvitationCodeEnabled
	s.PromoCodeEnabled = value.PromoCodeEnabled
}
