package requeststate

import (
	"errors"
	"strings"

	"github.com/tidwall/gjson"
)

// ValidateOpenAIReasoningEffort 拒绝 Codex 客户端专用的 Ultra 模式。
// Ultra 在 Codex 内部表示 max 推理加主动多代理，不是 OpenAI 上游协议档位。
func ValidateOpenAIReasoningEffort(body []byte, requestedModel string) error {
	efforts := []string{
		gjson.GetBytes(body, "reasoning.effort").String(),
		gjson.GetBytes(body, "reasoning_effort").String(),
		gjson.GetBytes(body, "output_config.effort").String(),
		gjson.GetBytes(body, "response.reasoning.effort").String(),
		gjson.GetBytes(body, "response.reasoning_effort").String(),
		gjson.GetBytes(body, "session.reasoning.effort").String(),
		gjson.GetBytes(body, "session.reasoning_effort").String(),
	}
	for _, effort := range efforts {
		if strings.EqualFold(strings.TrimSpace(effort), "ultra") {
			return errors.New(`reasoning effort "ultra" is not supported; use "max"`)
		}
	}

	return nil
}
