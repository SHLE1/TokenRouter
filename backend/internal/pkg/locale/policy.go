package locale

import (
	"encoding/json"
	"strconv"

	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
)

// PolicyMessages 保存规则中两种拦截提示及其稳定编辑标识。
type PolicyMessages struct {
	ID                               string          `json:"id,omitempty"`
	ErrorMessageLocalization         *Update[string] `json:"error_message_localization,omitempty"`
	FallbackErrorMessageLocalization *Update[string] `json:"fallback_error_message_localization,omitempty"`
}

// EnsureID 为历史规则生成编辑标识，首次保存后随条目持久化。
func (p *PolicyMessages) EnsureID(index int) {
	if p.ID == "" {
		p.ID = "legacy-" + strconv.Itoa(index)
	}
}

// PreparePolicyMessages 对规则原文和译文执行相同的版本检查。
func PreparePolicyMessages(old PolicyMessages, original, fallback string, input PolicyMessages) (PolicyMessages, error) {
	result := input
	pairs := []struct {
		old    *Update[string]
		next   **Update[string]
		source string
	}{
		{old.ErrorMessageLocalization, &result.ErrorMessageLocalization, original},
		{old.FallbackErrorMessageLocalization, &result.FallbackErrorMessageLocalization, fallback},
	}
	for _, pair := range pairs {
		if *pair.next == nil {
			*pair.next = pair.old
			continue
		}
		current := Original(pair.source)
		if pair.old != nil {
			current = pair.old.Content
		}
		next, err := Prepare(current, **pair.next, func(value string) error {
			if len(value) > 8192 {
				return apperror.BadRequest("LOCALIZED_TEXT_TOO_LONG", "Text exceeds the allowed length.")
			}
			return nil
		})
		if err != nil {
			return result, err
		}
		*pair.next = &Update[string]{Content: next}
	}
	return result, nil
}

// ResolvePolicyMessage 选择配置译文，留空时由策略提供内置提示。
func ResolvePolicyMessage(content *Update[string], original, language string) string {
	if content == nil {
		return original
	}
	value, _ := content.Resolve(language)
	return value
}

// ClonePolicySettings 复制规则及其嵌套译文，防止请求改写共享配置。
func ClonePolicySettings[T any](value *T) (*T, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
