package provider

import "strings"

// ResolveCompactForwardModel 应用 Compact 附加映射，空值时使用回退模型。
func ResolveCompactForwardModel(value *Record, model string) string {
	model = strings.TrimSpace(model)
	if model == "" || value == nil {
		return model
	}
	mapped, matched := value.ResolveCompactMappedModel(model)
	if matched && strings.TrimSpace(mapped) != "" {
		return strings.TrimSpace(mapped)
	}
	return model
}
