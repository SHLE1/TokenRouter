package provider

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/account"
	"github.com/TokenFlux/TokenRouter/internal/upstream/antigravity"
)

// resolveAntigravityModel 只执行一次映射，白名单在 thinking 规范化之后检查。
func resolveAntigravityModel(value *account.Record, requested string) string {
	if value == nil {
		return ""
	}
	requested = strings.TrimSpace(strings.TrimPrefix(requested, "models/"))
	if requested == "" {
		return ""
	}
	mapping := account.ResolveModelMapping(value, ModelDefaults())
	mapped, _ := account.ResolveMappedModel(value.Platform, mapping, requested)
	return mapped
}

// MapAntigravityModel 返回不改变 thinking 档位时的最终可用模型。
func MapAntigravityModel(value *account.Record, requested string) string {
	return FinalAntigravityModel(value, requested, nil)
}

// FinalAntigravityModel 在单跳映射和 thinking 规范化后检查最终白名单。
func FinalAntigravityModel(value *account.Record, requested string, thinking *bool) string {
	mapped := resolveAntigravityModel(value, requested)
	if mapped == "" {
		return ""
	}
	if thinking != nil {
		mapped = antigravity.ApplyThinkingModelSuffix(mapped, *thinking)
	}
	if !value.FinalModelWhitelisted(mapped, ModelDefaults(), ModelRules(value)) {
		return ""
	}
	return mapped
}
