package httpapi

import (
	"bytes"
	"context"
	"errors"

	"github.com/TokenFlux/TokenRouter/internal/creative"
	"github.com/TokenFlux/TokenRouter/internal/gateway/tierpolicy"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	identityhttp "github.com/TokenFlux/TokenRouter/internal/identity/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/payment"
	"github.com/TokenFlux/TokenRouter/internal/settings"
	"github.com/TokenFlux/TokenRouter/internal/settings/composite"
)

// CompositeSettings 提供综合设置读取和已登记的更新流程。
type CompositeSettings interface {
	GetAllSettings(context.Context) (*composite.Snapshot, error)
	GetAuthSourceDefaultSettings(context.Context) (*identity.AuthSourceDefaultSettings, error)
	GetOpenAIFastPolicySettings(context.Context) (*tierpolicy.OpenAIFastPolicySettings, error)
	IsTotpEncryptionKeyConfigured() bool
	OIDCSecurityWriteDefaults(context.Context) (bool, bool, error)
	BeginSettingsUpdate(context.Context) (*settings.UpdateSession, error)
	PrepareSettingsWithAuthSourceDefaults(context.Context, *composite.Snapshot, *identity.AuthSourceDefaultSettings, settings.OmittedKeys) (map[string]string, error)
	ApplyPersistedSettings(context.Context) error
}
type MonitoringSettings interface {
	IsMonitoringEnabled(context.Context) bool
	SetMonitoringEnabled(bool)
}
type PaymentSettings interface {
	GetPaymentConfig(context.Context) (*payment.PaymentConfig, error)
}
type CaptchaSecretValidator interface {
	ValidateSecretKey(context.Context, string) error
}
type CaptchaCredentialsValidator interface {
	ValidateCredentials(context.Context, string, string, string, string) error
}
type CreativeModelSettings interface {
	ListCreativeModelCandidates(context.Context) ([]creative.CreativeModelCandidate, error)
}

// Handler 仅拥有综合 HTTP 接口依赖，所有缓存与业务实例由 app 唯一装配。
type Handler struct {
	settingService       CompositeSettings
	settingsParticipants *settings.Registry
	participantError     error
	opsService           MonitoringSettings
	paymentConfigService PaymentSettings
	turnstileService     CaptchaSecretValidator
	aliyunCaptchaService CaptchaCredentialsValidator
	userAttributeService *identity.UserAttributeService
	totpService          identityhttp.StepUpGrantChecker
	userService          identityhttp.UserReader
	creativeModelReader  CreativeModelSettings
}
type HandlerOptions struct {
	Settings         CompositeSettings
	Participants     *settings.Registry
	ParticipantError error
	Monitoring       MonitoringSettings
	Payment          PaymentSettings
	Turnstile        CaptchaSecretValidator
	Aliyun           CaptchaCredentialsValidator
	Attributes       *identity.UserAttributeService
	Totp             identityhttp.StepUpGrantChecker
	User             identityhttp.UserReader
	Creative         CreativeModelSettings
}

// NewHandler 构造不回源、不启动任务；参与者必须在路由开放前固定。
func NewHandler(o HandlerOptions) *Handler {
	return &Handler{settingService: o.Settings, settingsParticipants: o.Participants, participantError: o.ParticipantError, opsService: o.Monitoring, paymentConfigService: o.Payment, turnstileService: o.Turnstile, aliyunCaptchaService: o.Aliyun, userAttributeService: o.Attributes, totpService: o.Totp, userService: o.User, creativeModelReader: o.Creative}
}

func (h *Handler) preparedParticipants(ctx context.Context, input settings.Fields, values map[string]string) ([]settings.PreparedChange, error) {
	if h.participantError != nil {
		return nil, h.participantError
	}
	current := values
	_, writingTextSettings := input["localized_settings"]
	if _, writing := input["site_texts"]; writing || writingTextSettings || bytes.Contains(input["openai_fast_policy_settings"], []byte("localization")) || len(input["login_agreement_documents"]) > 0 || len(input["custom_menu_items"]) > 0 || len(input["custom_endpoints"]) > 0 || len(input["footer_links"]) > 0 {
		reader, ok := h.settingService.(interface {
			ReadLocalizationValues(context.Context) (map[string]string, error)
		})
		if !ok {
			return nil, errors.New("settings reader does not support localization")
		}
		persisted, err := reader.ReadLocalizationValues(ctx)
		if err != nil {
			return nil, err
		}
		current = make(map[string]string, len(values))
		for key, value := range values {
			current[key] = value
		}
		for _, key := range settings.LocalizedTextFields {
			delete(current, key)
			if value, exists := persisted[key]; exists {
				current[key] = value
			}
			stored := key + "_localized"
			delete(current, stored)
			if value, exists := persisted[stored]; exists {
				current[stored] = value
			}
		}
		for _, key := range []string{"openai_fast_policy_settings", "custom_menu_items", "custom_endpoints", "footer_links", "login_agreement_documents", "site_texts", "site_name", "site_title", "site_subtitle", "contact_info", "doc_url", "home_content", "purchase_subscription_url", "footer_text"} {
			delete(current, key)
			if value, exists := persisted[key]; exists {
				current[key] = value
			}
		}
	}
	prepared, err := h.settingsParticipants.Prepare(ctx, input, current)
	if err != nil {
		return nil, err
	}
	for _, key := range h.settingsParticipants.OwnedKeys() {
		delete(values, key)
	}
	return prepared, nil
}
