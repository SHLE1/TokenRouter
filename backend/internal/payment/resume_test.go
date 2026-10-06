package payment

import (
	"net/url"
	"testing"
)

// TestCanonicalizeReturnURLRemovesUserParameters 检查用户参数清理和服务端订单参数生成。
func TestCanonicalizeReturnURLRemovesUserParameters(t *testing.T) {
	t.Parallel()

	const base = "https://site.example.com/payment/result"
	for _, suffix := range []string{
		"", "?", "?trade_status=TRADE_SUCCESS", "?trade_status=TRADE_SUCCESS#fragment",
		"?order_id=1&out_trade_no=attacker&resume_token=attacker&status=success&trade_status=TRADE_SUCCESS",
		"?from=checkout&nested=%26trade_status%3DTRADE_SUCCESS", "?trade_status=%ZZ",
	} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			canonical, err := CanonicalizeReturnURL(base+suffix, "site.example.com", "")
			if err != nil {
				t.Fatal(err)
			}
			if canonical != base {
				t.Fatalf("return URL = %q, want %q", canonical, base)
			}
			result, err := ResumeBuildPaymentReturnURL(canonical, 99, "ORDER123", "server-token")
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(result)
			if err != nil {
				t.Fatal(err)
			}
			query := parsed.Query()
			if len(query) != 4 || query.Get("order_id") != "99" || query.Get("out_trade_no") != "ORDER123" ||
				query.Get("resume_token") != "server-token" || query.Get("status") != "success" {
				t.Fatalf("unexpected result parameters: %v", query)
			}
		})
	}
}
