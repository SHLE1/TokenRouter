package gateway

import (
	"context"
	"encoding/json"

	"github.com/TokenFlux/TokenRouter/internal/gateway/tierpolicy"
	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/settings"
)

// FastSettingsParticipant 复用纯档位规则，省略或 null 均保持原配置。
func FastSettingsParticipant() settings.Participant {
	return settings.Participant{Module: "gateway", Fields: []string{"openai_fast_policy_settings"}, Keys: []string{SettingKeyOpenAIFastPolicySettings}, Prepare: func(_ context.Context, input settings.Fields, current map[string]string) (settings.PreparedChange, error) {
		raw, ok := input["openai_fast_policy_settings"]
		if !ok || string(raw) == "null" {
			return settings.PreparedChange{}, nil
		}
		var value tierpolicy.OpenAIFastPolicySettings
		if err := json.Unmarshal(raw, &value); err != nil {
			return settings.PreparedChange{}, apperror.BadRequest("", err.Error())
		}
		if err := prepareFastMessages(current[SettingKeyOpenAIFastPolicySettings], &value); err != nil {
			return settings.PreparedChange{}, err
		}
		prepared, err := tierpolicy.Prepare(&value)
		if err != nil {
			return settings.PreparedChange{}, apperror.BadRequest("", err.Error())
		}
		change := settings.PreparedChange{Values: map[string]string{SettingKeyOpenAIFastPolicySettings: prepared}}
		if hasPolicyTranslations(prepared) {
			change.Expected = map[string]*string{SettingKeyOpenAIFastPolicySettings: nil}
			if old, exists := current[SettingKeyOpenAIFastPolicySettings]; exists {
				change.Expected[SettingKeyOpenAIFastPolicySettings] = &old
			}
		}
		return change, nil
	}}
}
