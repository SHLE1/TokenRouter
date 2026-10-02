package app

import (
	"context"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/provider"

	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// provideCodexImporter 绑定提供商管理、备份查询和平台接口。
func provideCodexImporter(admin *provider.Admin, archive *provider.Archive, invalidator provider.TokenCacheInvalidator) *provider.CodexImporter {
	options := provider.CodexImportOptions{OAuthClientID: openai.ClientID, ValidatePrivateKey: func(value string) error { _, err := openai.ParseAgentIdentityPrivateKey(value); return err }}
	if invalidator != nil {
		options.Invalidate = func(ctx context.Context, value *provider.Record) error {
			return invalidator.InvalidateToken(ctx, value)
		}
	}
	options.Now = time.Now
	return provider.NewCodexImporter(admin, archive, options)
}
