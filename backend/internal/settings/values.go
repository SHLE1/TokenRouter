package settings

import "strings"

// StringOrDefault 返回键对应的非空字符串，缺键或空串时返回 fallback，空白字符按输入保留。
func StringOrDefault(values map[string]string, key, fallback string) string {
	if value, ok := values[key]; ok && value != "" {
		return value
	}
	return fallback
}

// IsExplicitFalse 兼容原持久值中的四种关闭表示。
func IsExplicitFalse(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "false", "0", "off", "disabled":
		return true
	default:
		return false
	}
}
