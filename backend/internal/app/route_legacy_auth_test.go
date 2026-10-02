package app

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/gateway/admission"

	identityhttp "github.com/TokenFlux/TokenRouter/internal/identity/httpapi"
	notificationhttp "github.com/TokenFlux/TokenRouter/internal/notification/httpapi"
	paymenthttp "github.com/TokenFlux/TokenRouter/internal/payment/httpapi"
	routinghttp "github.com/TokenFlux/TokenRouter/internal/routing/httpapi"
	sitehttp "github.com/TokenFlux/TokenRouter/internal/site/httpapi"

	servermiddleware "github.com/TokenFlux/TokenRouter/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes 注册认证路由。
// 身份、会话和权限检查的说明见下方文档。
// @project-doc docs/domains/identity_and_tenancy.md#authentication_boundaries
func RegisterAuthRoutes(
	v1 *gin.RouterGroup,
	h *routeTestHandlers,
	jwtAuth identityhttp.JWTAuthMiddleware,
	auditLog servermiddleware.AuditLogMiddleware,
	rateLimiter *servermiddleware.RateLimiter,
	settingService *admission.BackendMode,
	panelRateLimiter *servermiddleware.PanelRateLimiter,
) {
	guards := identityhttp.AuthRouteMiddleware{JWT: gin.HandlerFunc(jwtAuth), Audit: gin.HandlerFunc(auditLog), BackendAuth: identityhttp.BackendModeAuthGuard(legacyBackendModeReader(settingService)), BackendUser: identityhttp.BackendModeUserGuard(legacyBackendModeReader(settingService)), Panel: panelRateLimiter.Global(), Limit: func(key string, n int, window time.Duration) gin.HandlerFunc {
		return rateLimiter.LimitWithOptions(key, n, window, servermiddleware.RateLimitOptions{FailureMode: servermiddleware.RateLimitFailClose})
	}}
	identityhttp.RegisterAuthenticationRoutes(v1, h.Auth, h.Passkey, guards, func(group *gin.RouterGroup) { paymenthttp.RegisterWeChatAuthRoutes(group, h.Auth) })
	public := v1.Group("/settings")
	public.Use(panelRateLimiter.PublicIP())
	sitehttp.RegisterPublicSettingsRoutes(public, h.PublicSettings)
	notificationhttp.RegisterUnsubscribeRoute(public, h.Notification)
	routinghttp.RegisterPublicMarketplaceRoutes(v1, h.ModelMarketplace)
	identityhttp.RegisterSessionRoutes(v1, h.Auth, guards)
}

// legacyBackendModeReader 将 nil 指针转换为 nil 接口。
func legacyBackendModeReader(s *admission.BackendMode) identityhttp.BackendModeReader {
	if s == nil {
		return nil
	}
	return s
}
