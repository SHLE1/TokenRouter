package httpapi

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain 在测试进程启动时设置 Gin 测试模式。
func TestMain(m *testing.M) { gin.SetMode(gin.TestMode); os.Exit(m.Run()) }
