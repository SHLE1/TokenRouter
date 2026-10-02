package composite

import "github.com/TokenFlux/TokenRouter/internal/billing"

// ApplyBillingAdminReadSettings 将计费读取结果写入综合快照。
func (s *Snapshot) ApplyBillingAdminReadSettings(value *billing.AdminReadSettings) {
	s.BalanceIconSVG = value.BalanceIconSVG
	s.BalanceLowNotifyEnabled = value.BalanceLowNotifyEnabled
	s.BalanceLowNotifyRechargeURL = value.BalanceLowNotifyRechargeURL
	s.BalanceLowNotifyThreshold = value.BalanceLowNotifyThreshold
	s.BalanceUnitName = value.BalanceUnitName
	s.BalanceUnitSymbol = value.BalanceUnitSymbol
	s.DefaultBalance = value.DefaultBalance
	s.DefaultSubscriptions = value.DefaultSubscriptions
	s.ReasoningPointRMBUnitPrice = value.ReasoningPointRMBUnitPrice
	s.SubscriptionExpiryNotifyEnabled = value.SubscriptionExpiryNotifyEnabled
	s.USDExchangeRate = value.USDExchangeRate
}
