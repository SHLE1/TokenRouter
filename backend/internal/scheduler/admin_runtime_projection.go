package scheduler

import (
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/scheduler/policy"
)

// RuntimeSettingsFromAdmin 按字段解析管理设置，分别记录是否设置，并在缺省时使用对应默认值。
func RuntimeSettingsFromAdmin(value AdminSettings, defaults AdminDefaults) policy.RuntimeSettings {
	errorAlpha, _ := ParseAdvancedSchedulerAlphaOverride(value.AdvancedSchedulerEWMAErrorRateAlpha, defaults.Process.EwmaErrorRateAlpha)
	ttftAlpha, _ := ParseAdvancedSchedulerAlphaOverride(value.AdvancedSchedulerEWMATTFTAlpha, defaults.Process.EwmaTTFTAlpha)
	stickyTTFT, _ := ParseAdvancedSchedulerPositiveFloatOverride(value.AdvancedSchedulerStickyEscapeTTFTMs, defaults.Process.StickyEscape.TtftMs)
	stickyRate, _ := ParseAdvancedSchedulerRateOverride(value.AdvancedSchedulerStickyEscapeErrorRate, defaults.Process.StickyEscape.ErrorRate)
	return policy.RuntimeSettings{
		StickyWeightedEnabled: value.AdvancedSchedulerStickyWeightedEnabled, SubscriptionPriorityEnabled: value.AdvancedSchedulerSubscriptionPriorityEnabled, LbTopKOverride: ParsePositiveIntOverride(value.AdvancedSchedulerLBTopK),
		EwmaErrorRateAlpha: errorAlpha, EwmaErrorRateAlphaSet: strings.TrimSpace(value.AdvancedSchedulerEWMAErrorRateAlpha) != "", EwmaTTFTAlpha: ttftAlpha, EwmaTTFTAlphaSet: strings.TrimSpace(value.AdvancedSchedulerEWMATTFTAlpha) != "",
		StickyEscapeEnabled: value.AdvancedSchedulerStickyEscapeEnabled, StickyEscapeEnabledSet: value.AdvancedSchedulerStickyEscapeEnabledSet, StickyEscapeTTFTMs: stickyTTFT, StickyEscapeTTFTMsSet: strings.TrimSpace(value.AdvancedSchedulerStickyEscapeTTFTMs) != "", StickyEscapeErrorRate: stickyRate, StickyEscapeErrorRateSet: strings.TrimSpace(value.AdvancedSchedulerStickyEscapeErrorRate) != "", StickyEscape: policy.StickyEscapeConfig{Enabled: value.AdvancedSchedulerStickyEscapeEnabled, TtftMs: stickyTTFT, ErrorRate: stickyRate},
		WeightOverrides: ParseAdvancedSchedulerWeightOverrides(map[string]string{
			SettingKeyAdvancedSchedulerWeightErrorRate:        value.AdvancedSchedulerWeightErrorRate,
			SettingKeyAdvancedSchedulerWeightLoad:             value.AdvancedSchedulerWeightLoad,
			SettingKeyAdvancedSchedulerWeightPreviousResponse: value.AdvancedSchedulerWeightPreviousResponse,
			SettingKeyAdvancedSchedulerWeightPriority:         value.AdvancedSchedulerWeightPriority,
			SettingKeyAdvancedSchedulerWeightQueue:            value.AdvancedSchedulerWeightQueue,
			SettingKeyAdvancedSchedulerWeightQuotaHeadroom:    value.AdvancedSchedulerWeightQuotaHeadroom,
			SettingKeyAdvancedSchedulerWeightReset:            value.AdvancedSchedulerWeightReset,
			SettingKeyAdvancedSchedulerWeightSessionSticky:    value.AdvancedSchedulerWeightSessionSticky,
			SettingKeyAdvancedSchedulerWeightTTFT:             value.AdvancedSchedulerWeightTTFT,
		}),
	}
}
