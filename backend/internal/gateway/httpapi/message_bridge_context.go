package httpapi

import (
	"github.com/gin-gonic/gin"
)

const OpenAICompatMessagesBridgeContextKey = "openai_compat_messages_bridge"

func SetOpenAICompatMessagesBridgeContext(c *gin.Context, enabled bool) {
	if c == nil || !enabled {
		return
	}
	c.Set(OpenAICompatMessagesBridgeContextKey, true)
}

func IsOpenAICompatMessagesBridgeContext(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, ok := c.Get(OpenAICompatMessagesBridgeContextKey)
	if !ok {
		return false
	}
	enabled, ok := value.(bool)
	return ok && enabled
}
