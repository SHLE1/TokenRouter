package completion

import "strings"

// responseModelMismatch 比较最终出站模型和上游原始响应中的模型声明。
func responseModelMismatch(result *Result) *bool {
	if result == nil || strings.TrimSpace(result.UpstreamResponseModel) == "" {
		return nil
	}
	sent := strings.TrimSpace(result.UpstreamModel)
	if sent == "" {
		sent = strings.TrimSpace(result.Model)
	}
	mismatch := !strings.EqualFold(sent, strings.TrimSpace(result.UpstreamResponseModel))
	return &mismatch
}
