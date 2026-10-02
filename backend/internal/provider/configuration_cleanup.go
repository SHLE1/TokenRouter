package provider

import (
	"maps"
)

// DiscardDeprecatedProviderExtra 静默移除旧客户端可能继续提交的废弃提供商扩展键。
func DiscardDeprecatedProviderExtra(extra map[string]any) {
	NormalizeLegacyOpenAIProviderExtra(extra)
	DiscardDeprecatedExtra(extra)
}

// NormalizeDeprecatedProviderExtraUpdate 规范化用于整份替换的提供商 Extra。
// 返回的布尔值表示是否执行替换：空对象清空 Extra，仅含废弃键时跳过更新。
func NormalizeDeprecatedProviderExtraUpdate(extra map[string]any) (map[string]any, bool) {
	if extra == nil {
		return nil, false
	}
	normalized := maps.Clone(extra)
	DiscardDeprecatedProviderExtra(normalized)
	if len(extra) > 0 && len(normalized) == 0 {
		return nil, false
	}
	return normalized, true
}
