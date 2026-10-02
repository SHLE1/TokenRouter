package httpapi

import "github.com/gin-gonic/gin"

// GatewaySettingsEndpoints 声明设置的 HTTP 操作，对应模块实现设置规则。
type GatewaySettingsEndpoints interface {
	GetRectifierSettings(*gin.Context)
	UpdateRectifierSettings(*gin.Context)
	GetBetaPolicySettings(*gin.Context)
	UpdateBetaPolicySettings(*gin.Context)
}

// RegisterGatewaySettingsRoutes 在已经鉴权和审计的设置组中注册原路径。
func RegisterGatewaySettingsRoutes(adminSettings *gin.RouterGroup, endpoint GatewaySettingsEndpoints) {
	adminSettings.GET("/rectifier", endpoint.GetRectifierSettings)
	adminSettings.PUT("/rectifier", endpoint.UpdateRectifierSettings)
	adminSettings.GET("/beta-policy", endpoint.GetBetaPolicySettings)
	adminSettings.PUT("/beta-policy", endpoint.UpdateBetaPolicySettings)
}
