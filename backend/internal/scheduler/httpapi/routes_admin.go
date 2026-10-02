package httpapi

import "github.com/gin-gonic/gin"

// RegisterProviderDiagnostics 在提供商路径下注册评分诊断路由，分组权限由 app 安装。
func RegisterProviderDiagnostics(providers *gin.RouterGroup, endpoint *DiagnosticsHandler) {
	providers.GET("/:id/advanced-scheduler-score", endpoint.GetAdvancedSchedulerScore)
	providers.POST("/:id/advanced-scheduler-score/preview", endpoint.PreviewAdvancedSchedulerScore)
}
