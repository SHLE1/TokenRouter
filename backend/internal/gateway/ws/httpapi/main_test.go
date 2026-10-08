package httpapi

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain 在运行测试前统一设置 Gin 的测试模式。
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
