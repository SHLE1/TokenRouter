package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/identity"
	identityhttp "github.com/TokenFlux/TokenRouter/internal/identity/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/notification"
	"github.com/TokenFlux/TokenRouter/internal/promotion"
	promotionhttp "github.com/TokenFlux/TokenRouter/internal/promotion/httpapi"
)

// identityVerificationDelivery 转换身份模块生成的验证邮件事件。
type identityVerificationDelivery struct {
	mail       *notification.Mailer
	challenges *identity.EmailChallenges
}

func (d identityVerificationDelivery) SendNotifyVerification(ctx context.Context, n identity.NotifyVerificationNotice) error {
	return d.mail.SendNotifyVerification(ctx, n.UserID, n.Email, n.Code, n.Locale, n.SiteName)
}

func providePanelUserHTTP(users *identity.UserService, g *identityAuthGraph, mail *notification.Mailer, cache identity.EmailCache, challenges *identity.EmailChallenges) *identityhttp.UserHandler {
	return identityhttp.NewUserHandler(users, g.Core, identityVerificationDelivery{mail: mail, challenges: challenges}, cache)
}

func providePromotionUserHTTP(s *promotion.AffiliateService) *promotionhttp.UserHandler {
	return promotionhttp.NewUserHandler(s)
}

// GenerateVerifyCode 调用身份模块生成验证码，再交给投递流程。
func (d identityVerificationDelivery) GenerateVerifyCode() (string, error) {
	return d.challenges.GenerateVerifyCode()
}
