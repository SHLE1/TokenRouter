package requeststate

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/routing"

	"github.com/TokenFlux/TokenRouter/internal/protocol/openai"
	"github.com/tidwall/gjson"
)

// ExtractOpenAIReasoningEffortFromBody 读取请求体指定的推理档位。
func ExtractOpenAIReasoningEffortFromBody(body []byte) *string {
	reasoningEffort := strings.TrimSpace(gjson.GetBytes(body, "reasoning.effort").String())
	if reasoningEffort == "" {
		reasoningEffort = strings.TrimSpace(gjson.GetBytes(body, "reasoning_effort").String())
	}
	if reasoningEffort != "" {
		normalized := openai.NormalizeRecordedReasoningEffort(reasoningEffort)
		if normalized == "" {
			return nil
		}
		return &normalized
	}

	return nil
}

// CanonicalRequestedReasoningEffort 读取策略改写前的请求推理档位，包括 none，字段缺失时返回空字符串。
func CanonicalRequestedReasoningEffort(body []byte) *string {
	raw := strings.TrimSpace(gjson.GetBytes(body, "reasoning.effort").String())
	if raw == "" {
		raw = strings.TrimSpace(gjson.GetBytes(body, "reasoning_effort").String())
	}
	if raw == "" {
		raw = strings.TrimSpace(gjson.GetBytes(body, "output_config.effort").String())
	}
	if raw != "" {
		canonical := routing.NormalizeRequestedOpenAIReasoningEffort(raw)
		if canonical == "" {
			return nil
		}
		return &canonical
	}

	return nil
}

func ExtractOpenAIServiceTier(reqBody map[string]any) *string {
	if reqBody == nil {
		return nil
	}
	raw, ok := reqBody["service_tier"].(string)
	if !ok {
		return nil
	}
	return openai.NormalizeServiceTier(raw)
}

func ExtractOpenAIServiceTierFromBody(body []byte) *string {
	if len(body) == 0 {
		return nil
	}
	return openai.NormalizeServiceTier(gjson.GetBytes(body, "service_tier").String())
}

// ExtractOpenAIReasoningEffort 从 map 请求体读取指定的推理档位。
func ExtractOpenAIReasoningEffort(reqBody map[string]any) *string {
	if value, present := openai.GetOpenAIReasoningEffortFromReqBody(reqBody); present {
		if value == "" {
			return nil
		}
		return &value
	}

	return nil
}

// CanonicalRequestedReasoningEffortFromReqBody 是 map 形态请求体的同等入口。
func CanonicalRequestedReasoningEffortFromReqBody(reqBody map[string]any) *string {
	if reqBody == nil {
		return nil
	}
	raw := ""
	if reasoning, ok := reqBody["reasoning"].(map[string]any); ok {
		if effort, ok := reasoning["effort"].(string); ok {
			raw = strings.TrimSpace(effort)
		}
	}
	if raw == "" {
		if effort, ok := reqBody["reasoning_effort"].(string); ok {
			raw = strings.TrimSpace(effort)
		}
	}
	if raw == "" {
		if outputConfig, ok := reqBody["output_config"].(map[string]any); ok {
			if effort, ok := outputConfig["effort"].(string); ok {
				raw = strings.TrimSpace(effort)
			}
		}
	}
	if raw != "" {
		canonical := routing.NormalizeRequestedOpenAIReasoningEffort(raw)
		if canonical == "" {
			return nil
		}
		return &canonical
	}
	return nil
}
