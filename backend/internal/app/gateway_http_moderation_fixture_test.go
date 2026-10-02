package app

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/moderation"
	moderationadapter "github.com/TokenFlux/TokenRouter/internal/moderation/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// newHTTPModeration 组合审核用例与本地协议客户端，测试结束前等待后台任务完成。
func newHTTPModeration(t *testing.T, settings moderation.SettingRepository, repo moderation.ContentModerationRepository) *moderation.ContentModerationService {
	t.Helper()
	var background sync.WaitGroup
	t.Cleanup(background.Wait)
	core := moderation.NewContentModerationService(settings, repo, nil, nil, nil, nil, nil, moderation.Runtime{
		Audit:         moderationadapter.NewAuditClient(),
		SnapshotMedia: moderationadapter.SnapshotMedia,
		Background:    func(_ string, fn func()) { background.Go(fn) },
		CyberText:     openai.IsOpenAICyberWarningText,
		CyberPolicy:   openai.DetectOpenAICyberPolicy,
		ErrorMessage:  upstream.ExtractErrorMessage,
		MissingRow:    func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
		MissingUser:   func(err error) bool { return errors.Is(err, identity.ErrUserNotFound) },
	})
	t.Cleanup(func() {
		if err := core.Stop(); err != nil {
			t.Errorf("停止审核运行时: %v", err)
		}
	})
	return core
}
