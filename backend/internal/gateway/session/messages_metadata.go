package session

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/scheduler"
	"github.com/tidwall/gjson"
)

func MessagesMetadataSession(claudeSessionID, sessionHash, promptCacheKey, reqModel string, body []byte) (string, string) {
	// Anthropic metadata.user_id 和 X-Claude-Code-Session-Id 用于本地提供商粘性，后者比内容摘要更稳定。
	// 上游 GPT/Codex 的 prompt_cache_key 和 session_id 由 ForwardAsAnthropic 根据 cache_control 或完整消息摘要派生，
	// 使后续 turn 的缓存键随内容滚动。
	if promptCacheKey == "" {
		if claudeSessionID != "" {
			return currentSessionHash(claudeSessionID), promptCacheKey
		}
	}
	if sessionHash != "" {
		return sessionHash, promptCacheKey
	}
	if userID := strings.TrimSpace(gjson.GetBytes(body, "metadata.user_id").String()); userID != "" {
		seed := reqModel + "-" + userID
		sessionHash = currentSessionHash(seed)
	}
	return sessionHash, promptCacheKey
}

// currentSessionHash 按调度会话的格式计算哈希。
func currentSessionHash(seed string) string {
	current, _ := scheduler.DeriveSessionHashes(seed)
	return current
}
