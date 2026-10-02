package app

import (
	"github.com/TokenFlux/TokenRouter/internal/creative"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/ops"
	"github.com/TokenFlux/TokenRouter/internal/payment"
	"github.com/TokenFlux/TokenRouter/internal/settings"
	"github.com/TokenFlux/TokenRouter/internal/settings/composite"
	settingshttp "github.com/TokenFlux/TokenRouter/internal/settings/httpapi"
)

// provideCompositeSettingsHTTP 为综合设置 HTTP 入口绑定各领域组件和设置协调器。
func provideCompositeSettingsHTTP(source *composite.Runtime, registry *settings.Registry, monitor *ops.OpsService, pay *payment.ConfigService, turnstile *identity.TurnstileService, aliyun *identity.AliyunCaptchaService, attributes *identity.UserAttributeService, totp *identity.TotpService, users *identity.UserService, creative *creative.Public) *settingshttp.Handler {
	return settingshttp.NewHandler(settingshttp.HandlerOptions{Settings: source, Participants: registry, Monitoring: monitor, Payment: pay, Turnstile: turnstile, Aliyun: aliyun, Attributes: attributes, Totp: totp, User: users, Creative: creative})
}
