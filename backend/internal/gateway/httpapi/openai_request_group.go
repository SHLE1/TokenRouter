package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/gin-gonic/gin"
)

// openAIRequestGroup 读取当前请求保存的分组。
func openAIRequestGroup(c *gin.Context) *routing.Group {
	if c == nil {
		return nil
	}
	v, ok := c.Get("api_key")
	if !ok {
		return nil
	}
	key, ok := v.(*apikey.APIKey)
	if !ok || key == nil {
		return nil
	}
	return key.Group
}
