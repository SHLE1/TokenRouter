package moderation

import (
	"context"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
)

// blockMessageContent 为历史配置提供稳定的编辑版本，内置提示附带英文译文。
func blockMessageContent(cfg *ContentModerationConfig) locale.Content[string] {
	if cfg.BlockMessageLocalization != nil {
		return cfg.BlockMessageLocalization.Content
	}
	content := locale.Original(cfg.BlockMessage)
	content.Revision, content.SourceRevision = 1, 1
	if cfg.BlockMessage == defaultContentModerationBlockMessage {
		language := "zh-Hans"
		content.SourceLocale = &language
		content.Translations["en"] = locale.Translation[string]{Value: "This request was blocked by the site's content policy. Revise your input and try again.", SourceRevision: 1}
	}
	return content
}

// UserBlockMessage 按请求语言选择本地审核阻断提示。
func (cfg *ContentModerationConfig) UserBlockMessage(ctx context.Context) string {
	message, _ := blockMessageContent(cfg).Resolve(locale.FromContext(ctx))
	return message
}

// validateBlockMessage 对原文和译文使用相同的长度限制。
func validateBlockMessage(value string) error {
	if strings.TrimSpace(value) == "" || len(value) > 8192 {
		return apperror.BadRequest("INVALID_BLOCK_MESSAGE", "Block message is required and must not exceed 8192 bytes.")
	}
	return nil
}
