package composite

import "github.com/TokenFlux/TokenRouter/internal/payment"

// ApplyPaymentAdminReadSettings 将支付读取结果写入综合快照。
func (s *Snapshot) ApplyPaymentAdminReadSettings(value *payment.AdminReadSettings) {
	s.PaymentVisibleMethodAlipayEnabled = value.PaymentVisibleMethodAlipayEnabled
	s.PaymentVisibleMethodAlipaySource = value.PaymentVisibleMethodAlipaySource
	s.PaymentVisibleMethodWxpayEnabled = value.PaymentVisibleMethodWxpayEnabled
	s.PaymentVisibleMethodWxpaySource = value.PaymentVisibleMethodWxpaySource
}
