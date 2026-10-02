package openai

import "strings"

func NormalizeRecordedReasoningEffort(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}

	// 兼容客户端常见的 x-high / x_high / x high 写法。
	value = strings.NewReplacer("-", "", "_", "", " ", "").Replace(value)

	switch value {
	case "none", "minimal":
		return ""
	case "low", "medium", "high":
		return value
	case "xhigh", "extrahigh":
		return "xhigh"
	case "max":
		return value
	default:
		// 使用记录接受已知档位，未知值返回空字符串。
		return ""
	}
}
