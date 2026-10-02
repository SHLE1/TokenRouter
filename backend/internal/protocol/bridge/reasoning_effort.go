package bridge

import (
	"strings"
)

// isUltraReasoningEffort 识别 Codex 客户端专用的 Ultra 模式。
func isUltraReasoningEffort(effort string) bool {
	return strings.EqualFold(strings.TrimSpace(effort), "ultra")
}
