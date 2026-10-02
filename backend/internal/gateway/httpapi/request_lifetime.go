package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// requestLifetime 引用应用的请求活动屏障。
type requestLifetime struct{ enter func() (func(), error) }

// BindRequestActivity 只在构造图完成、开放 HTTP 之前调用。
func (r *requestLifetime) BindRequestActivity(enter func() (func(), error)) { r.enter = enter }

// beginRequest 登记请求活动，应用停止后拒绝请求。
func (r *requestLifetime) beginRequest(c *gin.Context, format string) (func(), bool) {
	if r.enter == nil {
		return func() {}, true
	}
	release, err := r.enter()
	if err == nil {
		return release, true
	}
	const message = "Service is shutting down"
	switch format {
	case "google":
		WriteGoogleError(c, http.StatusServiceUnavailable, message)
	case "anthropic":
		WriteAnthropicError(c, http.StatusServiceUnavailable, "api_error", "", message)
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"type": "api_error", "message": message}})
	}
	return nil, false
}
