package openai

// GetOpenAIReasoningEffortFromReqBody 提取请求字段指定的推理档位，字段缺失时返回空字符串。
func GetOpenAIReasoningEffortFromReqBody(reqBody map[string]any) (value string, present bool) {
	if reqBody == nil {
		return "", false
	}

	// 优先读取 reasoning.effort。
	if reasoning, ok := reqBody["reasoning"].(map[string]any); ok {
		if effort, ok := reasoning["effort"].(string); ok {
			return NormalizeRecordedReasoningEffort(effort), true
		}
	}

	// 部分客户端通过顶层字段传递。
	if effort, ok := reqBody["reasoning_effort"].(string); ok {
		return NormalizeRecordedReasoningEffort(effort), true
	}

	return "", false
}
