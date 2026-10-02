package upstream

import (
	"strings"
)

// IsHTMLResponse 按 doctype 或 html 标签前缀识别 HTML 响应。
func IsHTMLResponse(body []byte) bool {
	trimmed := strings.TrimSpace(strings.ToLower(string(body)))
	return strings.HasPrefix(trimmed, "<!doctype html") ||
		strings.HasPrefix(trimmed, "<html")
}
