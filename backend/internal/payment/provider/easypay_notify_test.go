package provider

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/payment"
)

// TestEasyPayNotificationRejectsCheckoutSignature 回放浏览器取得的下单签名。
func TestEasyPayNotificationRejectsCheckoutSignature(t *testing.T) {
	t.Parallel()

	e := &EasyPay{config: map[string]string{
		"pid": "1000", "pkey": "test-merchant-secret",
		"apiBase": "https://pay.example.com", "paymentMode": "popup",
		"notifyUrl": "https://site.example.com/api/v1/payment/webhook/easypay",
	}}
	const prefix = "https://site.example.com/payment/result?order_id=99&out_trade_no=ORDER123&status=success"
	created, err := e.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID: "ORDER123", PaymentType: payment.TypeAlipay,
		Subject: "balance recharge", Amount: "650.00",
		ReturnURL: prefix + "&trade_status=TRADE_SUCCESS",
	})
	if err != nil {
		t.Fatal(err)
	}
	payURL, err := url.Parse(created.PayURL)
	if err != nil {
		t.Fatal(err)
	}

	for _, forged := range []bool{false, true} {
		name := "checkout replay"
		if forged {
			name = "status extracted from return URL"
		}
		t.Run(name, func(t *testing.T) {
			callback := payURL.Query()
			if forged {
				callback.Set("return_url", prefix)
				callback.Set("trade_status", tradeStatusSuccess)
			}
			// 拆出状态字段后签名仍然匹配，拒绝请求依赖回调字段检查。
			params := make(map[string]string, len(callback))
			for key := range callback {
				params[key] = callback.Get(key)
			}
			if !easyPayVerifySign(params, e.config["pkey"], callback.Get("sign")) {
				t.Fatal("test payload must reuse a valid checkout signature")
			}
			notification, err := e.VerifyNotification(context.Background(), callback.Encode(), nil)
			if err == nil || notification != nil {
				t.Fatalf("checkout signature accepted: notification=%+v, err=%v", notification, err)
			}
		})
	}
}

// TestEasyPayNotificationParameters 覆盖标准通知、额外字段和重复字段。
func TestEasyPayNotificationParameters(t *testing.T) {
	t.Parallel()

	e := &EasyPay{config: map[string]string{"pid": "1000", "pkey": "test-merchant-secret"}}
	tests := []struct {
		name      string
		extraKey  string
		extra     string
		duplicate string
		wantError bool
	}{
		{name: "standard notification"},
		{name: "optional param", extraKey: "param", extra: "merchant=value&extra=中文"},
		{name: "empty optional param", extraKey: "param"},
		{name: "return URL", extraKey: "return_url", extra: "https://site.example.com/payment/result", wantError: true},
		{name: "notify URL", extraKey: "notify_url", extra: "https://site.example.com/notify", wantError: true},
		{name: "device", extraKey: "device", extra: "mobile", wantError: true},
		{name: "unknown field", extraKey: "unknown", extra: "value", wantError: true},
		{name: "empty unknown field", extraKey: "unknown", wantError: true},
		{name: "empty return URL", extraKey: "return_url", wantError: true},
		{name: "duplicate status", duplicate: "trade_status", wantError: true},
		{name: "duplicate order", duplicate: "out_trade_no", wantError: true},
		{name: "duplicate signature", duplicate: "sign", wantError: true},
		{name: "duplicate amount", duplicate: "money", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := map[string]string{
				"pid": "1000", "trade_no": "UPSTREAM123", "out_trade_no": "ORDER123",
				"type": "alipay", "name": "充值 & 套餐=10%", "money": "650.00",
				"trade_status": tradeStatusSuccess,
			}
			if tt.extraKey != "" {
				params[tt.extraKey] = tt.extra
			}
			params["sign"] = easyPaySign(params, e.config["pkey"])
			params["sign_type"] = signTypeMD5
			values := url.Values{}
			for key, value := range params {
				values.Set(key, value)
			}
			if tt.duplicate != "" {
				values.Add(tt.duplicate, "different-value")
			}
			raw := values.Encode()
			notification, err := e.VerifyNotification(context.Background(), raw, nil)
			if tt.wantError {
				if err == nil || notification != nil {
					t.Fatalf("invalid parameters accepted: notification=%+v, err=%v", notification, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if notification.Status != payment.ProviderStatusSuccess || notification.Amount != 650 ||
				notification.OrderID != "ORDER123" || notification.TradeNo != "UPSTREAM123" ||
				notification.Metadata["pid"] != "1000" || notification.RawData != raw {
				t.Fatalf("unexpected notification: %+v", notification)
			}
			// 同一份有效通知篡改金额后应当验签失败。
			values.Set("money", "651.00")
			if _, err := e.VerifyNotification(context.Background(), values.Encode(), nil); err == nil || !strings.Contains(err.Error(), "invalid signature") {
				t.Fatalf("tampered notification: %v", err)
			}
		})
	}
}
