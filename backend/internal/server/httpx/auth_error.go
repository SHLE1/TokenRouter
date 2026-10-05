package httpx

import (
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
	"github.com/gin-gonic/gin"
)

// ErrorResponse 标准错误响应结构
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// AbortWithError 中断请求并返回JSON错误
func AbortWithError(c *gin.Context, statusCode int, code, message string) {
	language := locale.Default()
	if c.Request != nil {
		language = locale.Negotiate(c.GetHeader("Accept-Language"), locale.FromContext(c.Request.Context()))
	}
	c.JSON(statusCode, NewErrorResponse(code, locale.ErrorText(language, code, statusCode, message)))
	c.Abort()
}
