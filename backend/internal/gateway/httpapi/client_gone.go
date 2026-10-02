package httpapi

import "github.com/gin-gonic/gin"

// FailoverClientGone 判断请求 context 是否因下游断开而取消，已断开时结束换号。
// 取消后重新选择提供商会返回 context.Canceled，容易误报为提供商耗尽的 502。在途的 detach 请求照常计费。
// 响应尚未提交时标记 499（client closed request），供访问日志归类。
func FailoverClientGone(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.Context().Err() == nil {
		return false
	}
	// 先停止 compact 心跳，在互斥锁下等待心跳写入结束，再标记响应状态。
	// 心跳已提交 200 时保留该状态码。
	if StopOpenAICompactSSEKeepaliveCommitted(c) {
		return true
	}
	if !c.Writer.Written() {
		c.Status(StatusClientClosedRequest)
	}
	return true
}
