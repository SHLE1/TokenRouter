package app

import (
	"github.com/TokenFlux/TokenRouter/internal/apikey"
	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// provideOpsObservationAccess 返回观测所需的身份数据，失败 Key 用于错误记录。
func provideOpsObservationAccess() gatewayhttp.OpsObservationAccess {
	return gatewayhttp.OpsObservationAccess{
		APIKey: func(c *gin.Context) *apikey.APIKey {
			if key, ok := gatewayhttp.EffectiveAPIKey(c); ok && key != nil {
				return key
			}
			key, _ := keyhttp.GetOpsFallbackAPIKey(c)
			return key
		},
		Rejected: func(c *gin.Context) bool {
			_, rejected := middleware.GetIngressRejectReason(c)
			return rejected
		},
	}
}
