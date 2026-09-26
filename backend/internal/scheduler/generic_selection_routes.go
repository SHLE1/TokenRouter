package scheduler

import (
	"context"
	"fmt"
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// 基础单平台与混合选择使用同一独立投影，保留原粘性及优先级/最近使用顺序。
func (s *GenericSelector) selectRoutes(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}) (*FlowAccount, error) {
	// 优先检查 context 中的强制平台（/antigravity 路由）
	var platform string
	var resolvedGroup *FlowGroup
	forcePlatform, hasForcePlatform := s.ports.ForcePlatform(ctx)
	if hasForcePlatform && forcePlatform != "" {
		platform = forcePlatform
		if groupID != nil {
			group, err := s.ports.ResolveGroupByID(ctx, *groupID)
			if err != nil {
				return nil, err
			}
			ctx = s.ports.WithGroupContext(ctx, group)
			resolvedGroup = group
		}
	} else if groupID != nil {
		group, resolvedGroupID, err := s.ports.ResolveGatewayGroup(ctx, groupID)
		if err != nil {
			return nil, err
		}
		groupID = resolvedGroupID
		ctx = s.ports.WithGroupContext(ctx, group)
		resolvedGroup = group
		platform = ""
	} else {
		// 无分组不能产生候选账号，后续查询返回空池。
		platform = ""
	}

	// count_tokens 与可用性探测不占并发槽，但高级分组仍必须复用与主请求相同的
	// 最终分组、硬过滤和评分逻辑，不能退回基础排序。
	if resolvedGroup != nil && resolvedGroup.UsesAdvancedScheduler() {
		selection, err := s.Select(WithSelectOnly(ctx), SelectionInput{GroupID: groupID, SessionHash: sessionHash, RequestedModel: requestedModel, ExcludedIDs: excludedIDs})
		if err != nil {
			return nil, err
		}
		if selection == nil || selection.Account == nil {
			return nil, ErrNoAvailableAccounts
		}
		if selection.ReleaseFunc != nil {
			selection.ReleaseFunc()
		}
		return selection.Account, nil
	}

	// 入口已经完成回退准入，模型检查使用当前授权分组。
	if s.ports.CheckGroupModelRestriction(ctx, groupID, requestedModel) {
		s.diagnostics.event("warn", "group model restriction blocked request",
			"group_id", derefGroupID(groupID),
			"model", requestedModel)
		return nil, fmt.Errorf("%w supporting model: %s (group model restriction)", ErrNoAvailableAccounts, requestedModel)
	}

	account, err := s.SelectPlatform(ctx, groupID, sessionHash, requestedModel, excludedIDs, platform)
	if err != nil {
		return nil, err
	}
	return s.ports.HydrateSelectedAccount(ctx, account)
}

func (s *GenericSelector) SelectPlatform(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, platform string) (*FlowAccount, error) {
	preferOAuth := platform == capability.PlatformGemini
	routingAccountIDs := s.ports.RoutingAccountIDsForRequest(ctx, groupID, requestedModel, platform)

	var schedGroup *FlowGroup
	if groupID != nil && s.ports.ReadGroup != nil {
		schedGroup, _ = s.ports.ReadGroup(ctx, *groupID)
	}
	// upstream 依据必须覆盖路由、粘性和普通候选的全部旧版选择分支。
	needsUpstreamCheck := s.ports.NeedsUpstreamGroupRestrictionCheck(ctx, groupID)
	isUpstreamAllowed := func(account *FlowAccount) bool {
		return !needsUpstreamCheck || !s.ports.IsUpstreamModelRestrictedByGroup(ctx, *groupID, account, requestedModel)
	}

	var accounts []FlowAccount
	accountsLoaded := false

	if len(routingAccountIDs) > 0 {
		if s.ports.DebugModelRoutingEnabled() {
			s.diagnostics.printf("service.gateway", "[ModelRoutingDebug] legacy routed begin: group_id=%v model=%s platform=%s session=%s routed_ids=%v",
				derefGroupID(groupID), requestedModel, platform, shortFlowSessionHash(sessionHash), routingAccountIDs)
		}

		if sessionHash != "" && s.cache != nil {
			accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
			if err == nil && accountID > 0 && slices.Contains(routingAccountIDs, accountID) {
				if _, excluded := excludedIDs[accountID]; !excluded {
					account, err := s.ports.GetSchedulableAccount(ctx, accountID)

					if err == nil {
						clearSticky := s.ports.ShouldClearStickySessionForAccountLayer(ctx, account, requestedModel)
						if clearSticky {
							_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
						}
						if !clearSticky && s.ports.IsAccountInGroup(account, groupID) && account.Platform == platform && (requestedModel == "" || s.ports.IsModelSupportedByAccountWithContext(ctx, account, requestedModel)) && isUpstreamAllowed(account) && s.ports.IsAccountSchedulableForModelSelection(ctx, account, requestedModel) && s.ports.IsAccountSchedulableForQuota(account) && s.ports.IsAccountSchedulableForWindowCost(ctx, account, true) && s.ports.IsAccountSchedulableForRPM(ctx, account, true) {
							if s.ports.DebugModelRoutingEnabled() {
								s.diagnostics.printf("service.gateway", "[ModelRoutingDebug] legacy routed sticky hit: group_id=%v model=%s session=%s account=%d", derefGroupID(groupID), requestedModel, shortFlowSessionHash(sessionHash), accountID)
							}
							return account, nil
						}
					}
				}
			}
		}

		forcePlatform, hasForcePlatform := s.ports.ForcePlatform(ctx)
		if hasForcePlatform && forcePlatform == "" {
			hasForcePlatform = false
		}
		var err error
		accounts, _, err = s.ports.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
		accountsLoaded = true

		ctx = s.ports.WithWindowCostPrefetch(ctx, accounts)
		ctx = s.ports.WithRPMPrefetch(ctx, accounts)

		routingSet := make(map[int64]struct{}, len(routingAccountIDs))
		for _, id := range routingAccountIDs {
			if id > 0 {
				routingSet[id] = struct{}{}
			}
		}

		var selected *FlowAccount
		for i := range accounts {
			acc := &accounts[i]
			if _, ok := routingSet[acc.ID]; !ok {
				continue
			}
			if _, excluded := excludedIDs[acc.ID]; excluded {
				continue
			}

			if !s.ports.IsAccountSchedulableForSelection(acc) {
				continue
			}

			if schedGroup != nil && schedGroup.RequirePrivacySet && !acc.IsPrivacySet() {
				_ = s.ports.SetAccountError(ctx, acc.ID,
					fmt.Sprintf("Privacy not set, required by group [%s]", schedGroup.Name))
				continue
			}
			if requestedModel != "" && !s.ports.IsModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
				continue
			}
			if !isUpstreamAllowed(acc) {
				continue
			}
			if !s.ports.IsAccountSchedulableForModelSelection(ctx, acc, requestedModel) {
				continue
			}
			if !s.ports.IsAccountSchedulableForQuota(acc) {
				continue
			}
			if !s.ports.IsAccountSchedulableForWindowCost(ctx, acc, false) {
				continue
			}
			if !s.ports.IsAccountSchedulableForRPM(ctx, acc, false) {
				continue
			}
			if selected == nil {
				selected = acc
				continue
			}
			if acc.Priority < selected.Priority {
				selected = acc
			} else if acc.Priority == selected.Priority {
				switch {
				case acc.LastUsedAt == nil && selected.LastUsedAt != nil:
					selected = acc
				case acc.LastUsedAt != nil && selected.LastUsedAt == nil:

				case acc.LastUsedAt == nil && selected.LastUsedAt == nil:
					if preferOAuth && acc.Type != selected.Type && acc.Type == capability.AccountTypeOAuth {
						selected = acc
					}
				default:
					if acc.LastUsedAt.Before(*selected.LastUsedAt) {
						selected = acc
					}
				}
			}
		}

		if selected != nil {
			if sessionHash != "" && s.cache != nil {
				if err := s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, selected.ID, stickySessionTTL); err != nil {
					s.diagnostics.printf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, selected.ID, err)
				}
			}
			if s.ports.DebugModelRoutingEnabled() {
				s.diagnostics.printf("service.gateway", "[ModelRoutingDebug] legacy routed select: group_id=%v model=%s session=%s account=%d", derefGroupID(groupID), requestedModel, shortFlowSessionHash(sessionHash), selected.ID)
			}
			return selected, nil
		}
		s.diagnostics.printf("service.gateway", "[ModelRouting] No routed accounts available for model=%s, falling back to normal selection", requestedModel)
	}

	if sessionHash != "" && s.cache != nil {
		accountID, err := s.cache.GetSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
		if err == nil && accountID > 0 {
			if _, excluded := excludedIDs[accountID]; !excluded {
				account, err := s.ports.GetSchedulableAccount(ctx, accountID)

				if err == nil {
					clearSticky := s.ports.ShouldClearStickySessionForAccountLayer(ctx, account, requestedModel)
					if clearSticky {
						_ = s.cache.DeleteSessionAccountID(ctx, derefGroupID(groupID), sessionHash)
					}
					if !clearSticky && s.ports.IsAccountInGroup(account, groupID) && account.Platform == platform && (requestedModel == "" || s.ports.IsModelSupportedByAccountWithContext(ctx, account, requestedModel)) && isUpstreamAllowed(account) && s.ports.IsAccountSchedulableForModelSelection(ctx, account, requestedModel) && s.ports.IsAccountSchedulableForQuota(account) && s.ports.IsAccountSchedulableForWindowCost(ctx, account, true) && s.ports.IsAccountSchedulableForRPM(ctx, account, true) {
						return account, nil
					}
				}
			}
		}
	}

	if !accountsLoaded {
		forcePlatform, hasForcePlatform := s.ports.ForcePlatform(ctx)
		if hasForcePlatform && forcePlatform == "" {
			hasForcePlatform = false
		}
		var err error
		accounts, _, err = s.ports.ListSchedulableAccounts(ctx, groupID, platform, hasForcePlatform)
		if err != nil {
			return nil, fmt.Errorf("query accounts failed: %w", err)
		}
	}

	ctx = s.ports.WithWindowCostPrefetch(ctx, accounts)
	ctx = s.ports.WithRPMPrefetch(ctx, accounts)

	var selected *FlowAccount
	for i := range accounts {
		acc := &accounts[i]
		if _, excluded := excludedIDs[acc.ID]; excluded {
			continue
		}

		if !s.ports.IsAccountSchedulableForSelection(acc) {
			continue
		}

		if schedGroup != nil && schedGroup.RequirePrivacySet && !acc.IsPrivacySet() {
			_ = s.ports.SetAccountError(ctx, acc.ID,
				fmt.Sprintf("Privacy not set, required by group [%s]", schedGroup.Name))
			continue
		}
		if requestedModel != "" && !s.ports.IsModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
			continue
		}
		if !isUpstreamAllowed(acc) {
			continue
		}
		if !s.ports.IsAccountSchedulableForModelSelection(ctx, acc, requestedModel) {
			continue
		}
		if !s.ports.IsAccountSchedulableForQuota(acc) {
			continue
		}
		if !s.ports.IsAccountSchedulableForWindowCost(ctx, acc, false) {
			continue
		}
		if !s.ports.IsAccountSchedulableForRPM(ctx, acc, false) {
			continue
		}
		if selected == nil {
			selected = acc
			continue
		}
		if acc.Priority < selected.Priority {
			selected = acc
		} else if acc.Priority == selected.Priority {
			switch {
			case acc.LastUsedAt == nil && selected.LastUsedAt != nil:
				selected = acc
			case acc.LastUsedAt != nil && selected.LastUsedAt == nil:

			case acc.LastUsedAt == nil && selected.LastUsedAt == nil:
				if preferOAuth && acc.Type != selected.Type && acc.Type == capability.AccountTypeOAuth {
					selected = acc
				}
			default:
				if acc.LastUsedAt.Before(*selected.LastUsedAt) {
					selected = acc
				}
			}
		}
	}

	if selected == nil {
		if err := s.ports.GroupModelUnsupportedErrorIfApplicable(ctx, accounts, requestedModel, platform, excludedIDs, false, groupID, schedGroup); err != nil {
			return nil, err
		}
		stats := s.ports.LogDetailedSelectionFailure(ctx, groupID, sessionHash, requestedModel, platform, accounts, excludedIDs, false)
		if requestedModel != "" {
			return nil, fmt.Errorf("%w supporting model: %s (%s)", ErrNoAvailableAccounts, requestedModel, stats)
		}
		return nil, ErrNoAvailableAccounts
	}

	if sessionHash != "" && s.cache != nil {
		if err := s.cache.SetSessionAccountID(ctx, derefGroupID(groupID), sessionHash, selected.ID, stickySessionTTL); err != nil {
			s.diagnostics.printf("service.gateway", "set session account failed: session=%s account_id=%d err=%v", sessionHash, selected.ID, err)
		}
	}

	return selected, nil
}

func (s *GenericSelector) SelectMixed(ctx context.Context, groupID *int64, sessionHash string, requestedModel string, excludedIDs map[int64]struct{}, nativePlatform string) (*FlowAccount, error) {
	return s.SelectPlatform(ctx, groupID, sessionHash, requestedModel, excludedIDs, "")
}
