package app

import (
	"github.com/TokenFlux/TokenRouter/internal/idempotency"
	"github.com/TokenFlux/TokenRouter/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// httpRouteSecurity 保存各路由族的 middleware 顺序，app 提供中间件实例。
type httpRouteSecurity struct {
	JWT, Admin, Audit, StepUp, BackendAuth, BackendUser gin.HandlerFunc
	Panel                                               *middleware.PanelRateLimiter
	AuthLimiter                                         *middleware.RateLimiter
}

type (
	authRouteMount    func(*gin.RouterGroup, httpRouteSecurity)
	userRouteMount    func(*gin.RouterGroup, httpRouteSecurity)
	adminRouteMount   func(*gin.RouterGroup, httpRouteSecurity, gin.HandlerFunc)
	gatewayRouteMount func(*gin.Engine)
	paymentRouteMount func(*gin.RouterGroup, httpRouteSecurity)
)

// provideHTTPRouteMount 组合各路由注册函数。
func provideHTTPRouteMount(auth authRouteMount, user userRouteMount, admin adminRouteMount, gateway gatewayRouteMount, payment paymentRouteMount,
	_ *idempotency.IdempotencyCoordinator, _ *idempotency.IdempotencyCleanupService,
) httpRouteMount {
	return func(r *gin.Engine, security httpRouteSecurity, catalog gin.HandlerFunc, pages func(*gin.RouterGroup)) {
		v1 := r.Group("/api/v1")
		auth(v1, security)
		user(v1, security)
		admin(v1, security, catalog)
		gateway(r)
		payment(v1, security)
		pages(v1)
	}
}
