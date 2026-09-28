package provider

// QuotaResult 额度获取结果
type QuotaResult struct {
	UsageInfo *UsageInfo     // 转换后的使用信息
	Raw       map[string]any // 原始响应，可存入 provider.Extra
}
