//go:build embed

package web

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
)

// HTMLCache manages the cached index.html with injected settings
type HTMLCache struct {
	mu              sync.RWMutex
	entries         map[string]CachedHTML
	baseHTMLHash    string // Hash of the original index.html (immutable after build)
	settingsVersion uint64 // Incremented when settings change
}

// CachedHTML represents the cache state
type CachedHTML struct {
	Content []byte
	ETag    string
}

// NewHTMLCache creates a new HTML cache instance
func NewHTMLCache() *HTMLCache {
	return &HTMLCache{}
}

// SetBaseHTML initializes the cache with the base HTML template
func (c *HTMLCache) SetBaseHTML(baseHTML []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := sha256.Sum256(baseHTML)
	c.baseHTMLHash = hex.EncodeToString(hash[:8]) // First 8 bytes for brevity
}

// Invalidate marks the cache as stale
func (c *HTMLCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = nil
	c.settingsVersion++
}

// Get 返回当前渲染快照。
func (c *HTMLCache) Get() *CachedHTML {
	cached, _ := c.Snapshot()
	return cached
}

// Snapshot 同时取得内容和失效代次，使后续回源只能发布到原代次。
func (c *HTMLCache) Snapshot() (*CachedHTML, uint64) {
	return c.SnapshotForLocale(locale.Default())
}

// SnapshotForLocale 在同一次锁保护内读取语言快照和失效版本。
func (c *HTMLCache) SnapshotForLocale(code string) (*CachedHTML, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot, ok := c.entries[code]
	if !ok {
		return nil, c.settingsVersion
	}
	return &CachedHTML{Content: append([]byte(nil), snapshot.Content...), ETag: snapshot.ETag}, c.settingsVersion
}

// Publish 保存默认语言的 HTML，供已有调用方使用。
func (c *HTMLCache) Publish(version uint64, html, settingsJSON []byte) CachedHTML {
	return c.PublishForLocale(locale.Default(), version, html, settingsJSON)
}

// PublishForLocale 将语言加入 ETag，跨过失效点的渲染保持为当前请求私有。
func (c *HTMLCache) PublishForLocale(code string, version uint64, html, settingsJSON []byte) CachedHTML {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := append([]byte(code+"\n"), settingsJSON...)
	rendered := CachedHTML{Content: html, ETag: c.generateETag(key)}
	if version == c.settingsVersion {
		if c.entries == nil {
			c.entries = map[string]CachedHTML{}
		}
		c.entries[code] = CachedHTML{Content: append([]byte(nil), html...), ETag: rendered.ETag}
	}
	return rendered
}

// generateETag creates an ETag from base HTML hash + settings hash
func (c *HTMLCache) generateETag(settingsJSON []byte) string {
	settingsHash := sha256.Sum256(settingsJSON)
	return `"` + c.baseHTMLHash + "-" + hex.EncodeToString(settingsHash[:8]) + `"`
}
