package completion

import "strings"

// responseModelMismatch 只比较最终出站模型和原始响应声明，不改变计费模型。
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
