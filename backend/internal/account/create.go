// 本文件维护 account 的所属能力；兼容入口复用唯一实现。
package account

import (
	"context"
	"maps"

	"github.com/TokenFlux/TokenRouter/internal/egress"
)

func (s *Admin) CreateAccount(ctx context.Context, input *CreateAccountInput) (*Record, error) {
	accountExtra := maps.Clone(input.Extra)
	DiscardDeprecatedAccountExtra(accountExtra)
	if err := NormalizeUpstreamUsageExtra(accountExtra); err != nil {
		return nil, err
	}
	accountExtra, err := NormalizeGrokMediaEligibilityExtra(input.Platform, accountExtra)
	if err != nil {
		return nil, err
	}
	if err := ValidateUpstreamRequestIDHeaderExtra(accountExtra); err != nil {
		return nil, err
	}

	// 绑定分组
	groupIDs := input.GroupIDs

	// 校验并规范化请求头覆写配置（header 名小写化、格式检查）
	if err := egress.NormalizeHeaderOverrideCredentials(input.Credentials); err != nil {
		return nil, err
	}
	// OAuth 兑换后不得持久化临时 SSO 或密码。
	input.Credentials = SanitizeStoredCredentials(input.Platform, input.Credentials)

	account, err := BuildAccountForCreate(input, accountExtra, s.options.Creation)
	if err != nil {
		return nil, err
	}
	// 只有新建账号需要生成并持久化机器身份；编辑旧账号时必须保留兼容回退语义。
	if account.IsQoderCosy() {
		s.options.Credentials.Prepare(account)
	}
	s.attachProxyForValidation(ctx, account)
	if err := s.options.Credentials.Validate(ctx, account); err != nil {
		return nil, err
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	// 绑定分组
	if len(groupIDs) > 0 {
		if err := s.accountRepo.BindGroups(ctx, account.ID, groupIDs); err != nil {
			return nil, err
		}
	}

	// 后置任务使用自己的账号值，不与返回给 HTTP 的可变对象共享。
	if account.Type == AccountTypeOAuth && (account.Platform == PlatformOpenAI || account.Platform == PlatformAntigravity) {
		value := CloneRecord(account)
		s.options.Background("service/admin_account.go:CreateAccount", func() {
			defer func() {
				if r := recover(); r != nil {
					event := "create_account_openai_privacy_panic"
					if value.Platform == PlatformAntigravity {
						event = "create_account_antigravity_privacy_panic"
					}
					s.options.Error(event, "account_id", value.ID, "recover", r)
				}
			}()
			if value.Platform == PlatformOpenAI {
				s.options.Privacy.EnsureOpenAIPrivacy(context.Background(), value)
			} else {
				s.options.Privacy.EnsureAntigravityPrivacy(context.Background(), value)
			}
		})
	}

	return account, nil
}

func (s *Admin) attachProxyForValidation(ctx context.Context, value *Record) {
	if s.options.Proxies == nil || value == nil || value.Proxy != nil || value.ProxyID == nil || *value.ProxyID <= 0 {
		return
	}
	if proxy, err := s.options.Proxies.GetByID(ctx, *value.ProxyID); err == nil && proxy != nil {
		value.Proxy = proxy
	}
}
