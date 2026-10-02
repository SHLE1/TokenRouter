package site

import (
	"context"
	"encoding/json"

	"github.com/TokenFlux/TokenRouter/internal/settings"
)

// SettingsParticipant 准备综合入口传入的站点设置，提交后使公开 HTML 缓存失效。
func SettingsParticipant() settings.Participant {
	keys := []string{SettingKeyAPIBaseURL, SettingKeyContactInfo, SettingKeyCustomEndpoints, SettingKeyCustomMenuItems, SettingKeyDocURL, SettingKeyFooterLinks, SettingKeyFooterText, SettingKeyFrontendURL, SettingKeyHideCcsImportButton, SettingKeyHomeContent, SettingKeyHomeFeaturedModels, SettingKeyLoginAgreementDocuments, SettingKeyLoginAgreementEnabled, SettingKeyLoginAgreementMode, SettingKeyLoginAgreementUpdatedAt, SettingKeyPurchaseSubscriptionEnabled, SettingKeyPurchaseSubscriptionURL, SettingKeySiteLogo, SettingKeySiteName, SettingKeySiteNameEn, SettingKeySiteNameZh, SettingKeySiteSubtitle, SettingKeySiteSubtitleEn, SettingKeySiteSubtitleZh, SettingKeySiteTitleEn, SettingKeySiteTitleZh, SettingKeyTableDefaultPageSize, SettingKeyTablePageSizeOptions}
	return settings.Participant{Module: "site", Fields: keys, Keys: keys, Prepare: func(_ context.Context, input settings.Fields, _ map[string]string) (settings.PreparedChange, error) {
		if len(input) == 0 {
			return settings.PreparedChange{}, nil
		}
		raw, err := json.Marshal(input)
		if err != nil {
			return settings.PreparedChange{}, err
		}
		var value AdminSettings
		if err = json.Unmarshal(raw, &value); err != nil {
			return settings.PreparedChange{}, err
		}
		values, err := PrepareAdminSettings(&value)
		if err != nil {
			return settings.PreparedChange{}, err
		}
		for key := range values {
			if _, ok := input[key]; !ok {
				delete(values, key)
			}
		}
		return settings.PreparedChange{Values: values}, nil
	}}
}
