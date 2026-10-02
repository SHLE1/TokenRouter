package middleware

import (
	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	"github.com/gin-gonic/gin"
)

// RequireGroupAssignment 将测试 context 中的 Key 数据传给网关分组门禁。
func RequireGroupAssignment(writeError gatewayhttp.GatewayErrorWriter) gin.HandlerFunc {
	return gatewayhttp.RequireGroupAssignment(gatewayhttp.GroupAssignmentOptions{Access: func(c *gin.Context) gatewayhttp.GroupAssignmentAccess {
		key, ok := keyhttp.GetAPIKeyFromContext(c)
		if !ok || key == nil {
			return gatewayhttp.GroupAssignmentAccess{}
		}
		_, noGroup := c.Get(gatewayhttp.CompositeKeyNoGroupContextKey)
		return gatewayhttp.GroupAssignmentAccess{Loaded: true, Assigned: key.GroupID != nil, CompositeNoGroup: key.IsComposite && noGroup}
	}, WriteError: writeError, Rejected: func(c *gin.Context) {
		gatewayhttp.MarkOpsClientBusinessLimited(c, gatewayhttp.OpsClientBusinessLimitedReasonAPIKeyGroupUnassigned)
		MarkIngressRejected(c, IngressRejectGroupUnassigned)
	}})
}
