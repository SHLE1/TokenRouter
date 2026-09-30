package notification

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/notification/contract"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

type SettingRepository = settings.Repository

var ErrSettingNotFound = settings.ErrSettingNotFound

type (
	NotificationEmailSendInput = SendRequest
	Sender                     interface {
		SendEmail(context.Context, string, string, string) error
	}
)

type (
	SMTPConfig    = contract.SMTPConfig
	SMTPTransport interface {
		Send(context.Context, *SMTPConfig, string, string, string) error
		Test(context.Context, *SMTPConfig) error
	}
)

const (
	defaultSiteName       = "TokenRouter"
	SettingKeySiteName    = "site_name"
	SettingKeyAPIBaseURL  = "api_base_url"
	SettingKeyFrontendURL = "frontend_url"
)
