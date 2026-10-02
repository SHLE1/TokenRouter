package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/ops"
	"github.com/gin-gonic/gin"
)

// OpsErrorLogQueue 只接收已冻结的观测值，队列生命周期由 app 管理。
type OpsErrorLogQueue interface {
	Enqueue(*ops.OpsService, *ops.OpsInsertErrorLogInput)
}

// OpsObservationAccess 为错误日志提供身份数据和准入拒绝信息。
type OpsObservationAccess struct {
	APIKey   func(*gin.Context) *apikey.APIKey
	Rejected func(*gin.Context) bool
}

func (a OpsObservationAccess) key(c *gin.Context) *apikey.APIKey {
	if a.APIKey == nil {
		return nil
	}
	return a.APIKey(c)
}

func (a OpsObservationAccess) rejected(c *gin.Context) bool {
	return a.Rejected != nil && a.Rejected(c)
}
