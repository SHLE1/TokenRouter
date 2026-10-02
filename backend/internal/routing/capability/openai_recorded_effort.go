package capability

import (
	protocolopenai "github.com/TokenFlux/TokenRouter/internal/protocol/openai"
)

// NormalizeRecordedOpenAIEffortForModel 按实际模型的 max 支持情况归一化记录值。
func NormalizeRecordedOpenAIEffortForModel(raw string, model string) string {
	value := protocolopenai.NormalizeRecordedReasoningEffort(raw)
	switch value {
	case "max":
		if !OpenAIModelSupportsReasoningEffort(model, value) {
			return ""
		}
	}
	return value
}
