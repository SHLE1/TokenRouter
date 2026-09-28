package middleware

import "github.com/gin-gonic/gin"

// invalidAuthClientKey 测试保留旧局部函数名。
func invalidAuthClientKey(c *gin.Context) string { return InvalidAuthClientKey(c) }
