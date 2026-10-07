package antigravity

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/protocol/anthropic"
)

// 通用 wire 类型只保留别名；默认模型、安全配置及 v1internal 包装仍由平台拥有。

type ClaudeRequest = anthropic.ClaudeRequest

type ClaudeMessage = anthropic.ClaudeMessage

type ThinkingConfig = anthropic.ThinkingConfig

type ClaudeTool = anthropic.ClaudeTool

type ClaudeCustomToolSpec = anthropic.ClaudeCustomToolSpec

type SystemBlock = anthropic.SystemBlock

type ClaudeResponse = anthropic.ClaudeResponse

type ClaudeUsage = anthropic.ClaudeUsage

type ClaudeError = anthropic.ClaudeError

// IsGeminiReasoningModel 标记需要省略强制工具参数的上游型号。
func IsGeminiReasoningModel(modelID string) bool {
	lower := strings.ToLower(modelID)
	for _, id := range []string{"gemini-2.5-flash-thinking", "gemini-3-pro-high", "gemini-3.1-pro-high", "gemini-3.6-flash-high", "gemini-3.6-flash-low", "gemini-3.6-flash-medium", "gemini-3.6-flash-tiered"} {
		if strings.Contains(lower, id) {
			return true
		}
	}
	return false
}
