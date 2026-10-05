package errorpolicy

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
)

// MessageLocalization 保存错误规则的原文和译文。
type MessageLocalization = locale.TextContent

// DisplayMessage 返回管理员配置的当前语言消息。
func (r *ErrorPassthroughRule) DisplayMessage(language string) string {
	if r.MessageLocalization.Revision > 0 {
		message, _ := locale.Content[string](r.MessageLocalization).Resolve(language)
		return message
	}
	if r.CustomMessage != nil {
		return *r.CustomMessage
	}
	return ""
}

// prepareMessage 在规则匹配配置保存前校验译文。
func prepareMessage(current *ErrorPassthroughRule, next *ErrorPassthroughRule) error {
	if next.MessageUpdate == nil {
		return nil
	}
	original := locale.Original("")
	if current != nil {
		original = locale.Content[string](current.MessageLocalization)
		if original.Revision == 0 && current.CustomMessage != nil {
			original = locale.Original(*current.CustomMessage)
		}
	}
	content, err := locale.Prepare(original, *next.MessageUpdate, func(value string) error {
		if strings.TrimSpace(value) == "" || len(value) > 8192 {
			return apperror.BadRequest("INVALID_CUSTOM_MESSAGE", "Custom message is required and must not exceed 8192 bytes.")
		}
		return nil
	})
	if err != nil {
		return err
	}
	next.MessageLocalization = MessageLocalization(content)
	next.CustomMessage = &content.Source
	next.MessageUpdate = nil
	return nil
}
