//go:build embed

package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHTMLLanguageIsolation 检查不同语言的内容、ETag 和失效代次彼此独立。
func TestHTMLLanguageIsolation(t *testing.T) {
	cache := NewHTMLCache()
	cache.SetBaseHTML([]byte("<html></html>"))
	_, version := cache.SnapshotForLocale("en")
	english := cache.PublishForLocale("en", version, []byte("English"), []byte(`{}`))
	chinese := cache.PublishForLocale("zh-Hans", version, []byte("中文"), []byte(`{}`))
	require.NotEqual(t, english.ETag, chinese.ETag)
	saved, _ := cache.SnapshotForLocale("en")
	require.Equal(t, "English", string(saved.Content))
	saved.Content[0] = 'X'
	saved, _ = cache.SnapshotForLocale("en")
	require.Equal(t, "English", string(saved.Content))
	cache.Invalidate()
	cache.PublishForLocale("en", version, []byte("Late"), []byte(`{}`))
	saved, _ = cache.SnapshotForLocale("en")
	require.Nil(t, saved)
	saved, _ = cache.SnapshotForLocale("zh-Hans")
	require.Nil(t, saved)
}
