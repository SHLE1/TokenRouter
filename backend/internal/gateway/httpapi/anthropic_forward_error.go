package httpapi

import (
	"github.com/gin-gonic/gin"
)

// AnthropicForwardErrorOutput 写出 Messages 错误，调用方管理响应提交和流状态。
type AnthropicForwardErrorOutput struct{ Context *gin.Context }

func (o AnthropicForwardErrorOutput) Message(status int, kind, message string) {
	o.Context.JSON(status, gin.H{"type": "error", "error": gin.H{"type": kind, "message": message}})
}

func (o AnthropicForwardErrorOutput) Raw(status int, body []byte) {
	o.Context.Data(status, "application/json", body)
}
