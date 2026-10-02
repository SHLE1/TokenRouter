package payment

// AdminReadSettings 包含支付方式的展示开关和来源。
type AdminReadSettings struct {
	PaymentVisibleMethodAlipayEnabled bool
	PaymentVisibleMethodAlipaySource  string
	PaymentVisibleMethodWxpayEnabled  bool
	PaymentVisibleMethodWxpaySource   string
}

// ReadAdminSettings 从传入的设置值解析支付方式的展示开关和来源。
func ReadAdminSettings(settings map[string]string) *AdminReadSettings {
	result := &AdminReadSettings{}

	result.PaymentVisibleMethodAlipaySource = NormalizeVisibleMethodSource("alipay", settings[SettingPaymentVisibleMethodAlipaySource])
	result.PaymentVisibleMethodWxpaySource = NormalizeVisibleMethodSource("wxpay", settings[SettingPaymentVisibleMethodWxpaySource])
	result.PaymentVisibleMethodAlipayEnabled = settings[SettingPaymentVisibleMethodAlipayEnabled] == "true"
	result.PaymentVisibleMethodWxpayEnabled = settings[SettingPaymentVisibleMethodWxpayEnabled] == "true"
	return result
}
