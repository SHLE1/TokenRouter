package gateway

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/TokenFlux/TokenRouter/internal/gateway/tierpolicy"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
)

// prepareFastMessages 将译文按规则 ID 与持久版本对应。
func prepareFastMessages(current string, value *tierpolicy.OpenAIFastPolicySettings) error {
	var old tierpolicy.OpenAIFastPolicySettings
	_ = json.Unmarshal([]byte(current), &old)
	byID := map[string]tierpolicy.OpenAIFastPolicyRule{}
	for i, rule := range old.Rules {
		rule.EnsureID(i)
		byID[rule.ID] = rule
	}
	seen := map[string]bool{}
	for i := range value.Rules {
		rule := &value.Rules[i]
		rule.EnsureID(i)
		if seen[rule.ID] {
			return locale.ErrConflict
		}
		seen[rule.ID] = true
		prior := byID[rule.ID]
		next, err := locale.PreparePolicyMessages(prior.PolicyMessages, prior.ErrorMessage, prior.FallbackErrorMessage, rule.PolicyMessages)
		if err != nil {
			return err
		}
		rule.PolicyMessages = next
		if next.ErrorMessageLocalization != nil {
			rule.ErrorMessage = next.ErrorMessageLocalization.Source
		}
		if next.FallbackErrorMessageLocalization != nil {
			rule.FallbackErrorMessage = next.FallbackErrorMessageLocalization.Source
		}
	}
	return nil
}

// prepareBetaMessages 将 Beta 规则的两个提示保存为独立内容块。
func prepareBetaMessages(current string, value *BetaPolicySettings) error {
	var old BetaPolicySettings
	_ = json.Unmarshal([]byte(current), &old)
	byID := map[string]BetaPolicyRule{}
	for i, rule := range old.Rules {
		rule.EnsureID(i)
		byID[rule.ID] = rule
	}
	seen := map[string]bool{}
	for i := range value.Rules {
		rule := &value.Rules[i]
		rule.EnsureID(i)
		if seen[rule.ID] {
			return locale.ErrConflict
		}
		seen[rule.ID] = true
		prior := byID[rule.ID]
		next, err := locale.PreparePolicyMessages(prior.PolicyMessages, prior.ErrorMessage, prior.FallbackErrorMessage, rule.PolicyMessages)
		if err != nil {
			return err
		}
		rule.PolicyMessages = next
		if next.ErrorMessageLocalization != nil {
			rule.ErrorMessage = next.ErrorMessageLocalization.Source
		}
		if next.FallbackErrorMessageLocalization != nil {
			rule.FallbackErrorMessage = next.FallbackErrorMessageLocalization.Source
		}
	}
	return nil
}

// saveLocalizedPolicy 在数据库事务中比较整份规则配置，并发编辑返回冲突。
func (s *RuntimeSettings) saveLocalizedPolicy(ctx context.Context, key, value, old string, existed bool) error {
	if writer, ok := s.settingRepo.(interface {
		CompareAndSetMultiple(context.Context, map[string]string, map[string]*string) error
	}); ok {
		var expected *string
		if existed {
			expected = &old
		}
		return writer.CompareAndSetMultiple(ctx, map[string]string{key: value}, map[string]*string{key: expected})
	}
	if hasPolicyTranslations(value) {
		return errors.New("settings repository does not support conditional updates")
	}
	return s.settingRepo.Set(ctx, key, value)
}

// hasPolicyTranslations 用于识别需要版本保护的规则写入。
func hasPolicyTranslations(raw string) bool {
	var value struct {
		Rules []locale.PolicyMessages `json:"rules"`
	}
	if json.Unmarshal([]byte(raw), &value) != nil {
		return false
	}
	for _, rule := range value.Rules {
		if rule.ErrorMessageLocalization != nil || rule.FallbackErrorMessageLocalization != nil {
			return true
		}
	}
	return false
}

// GetBetaPolicySettings 返回独立的规则副本，用户请求按当前语言选择提示。
func (s *RuntimeSettings) GetBetaPolicySettings(ctx context.Context) (*BetaPolicySettings, error) {
	stored, err := s.readBetaPolicySettings(ctx)
	if err != nil {
		return nil, err
	}
	value, err := locale.ClonePolicySettings(stored)
	if err != nil {
		return nil, err
	}
	for i := range value.Rules {
		rule := &value.Rules[i]
		rule.EnsureID(i)
		if locale.UserPresentation(ctx) {
			rule.ErrorMessage = locale.ResolvePolicyMessage(rule.ErrorMessageLocalization, rule.ErrorMessage, locale.FromContext(ctx))
			rule.FallbackErrorMessage = locale.ResolvePolicyMessage(rule.FallbackErrorMessageLocalization, rule.FallbackErrorMessage, locale.FromContext(ctx))
		}
	}
	return value, nil
}

// GetOpenAIFastPolicySettings 为 HTTP 和 WS 提供相同语言的策略提示。
func (s *RuntimeSettings) GetOpenAIFastPolicySettings(ctx context.Context) (*OpenAIFastPolicySettings, error) {
	stored, err := s.readOpenAIFastPolicySettings(ctx)
	if err != nil {
		return nil, err
	}
	value, err := locale.ClonePolicySettings(stored)
	if err != nil {
		return nil, err
	}
	for i := range value.Rules {
		rule := &value.Rules[i]
		rule.EnsureID(i)
		if locale.UserPresentation(ctx) {
			rule.ErrorMessage = locale.ResolvePolicyMessage(rule.ErrorMessageLocalization, rule.ErrorMessage, locale.FromContext(ctx))
			rule.FallbackErrorMessage = locale.ResolvePolicyMessage(rule.FallbackErrorMessageLocalization, rule.FallbackErrorMessage, locale.FromContext(ctx))
		}
	}
	return value, nil
}
