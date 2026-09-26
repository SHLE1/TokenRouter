package selection

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/gateway/media"

	"github.com/TokenFlux/TokenRouter/internal/account"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

type (
	candidatePolicyKey    struct{}
	imageModelRequiredKey struct{}
	preferredAccountsKey  struct{}
	candidatePolicy       struct {
		groupID  *int64
		stickyID int64
	}
)

// withCandidatePolicy 固定本次分组和粘性账号，供窗口费用、RPM 与重检共用。
func (s *Compatible) withCandidatePolicy(ctx context.Context, groupID *int64, session string) context.Context {
	state := candidatePolicy{groupID: groupID}
	if session != "" && s.cache != nil {
		state.stickyID, _ = s.getStickySessionAccountID(ctx, groupID, session)
	}
	return context.WithValue(ctx, candidatePolicyKey{}, state)
}

// candidateEligibilityReason 在协议和模型过滤后复用各平台已有的账号硬限制。
func (s *Compatible) candidateEligibilityReason(ctx context.Context, value *provider.ExecutionAccount, platform, model string, compact bool, required account.OpenAIEndpointCapability) string {
	if reason := provider.CompatibleEligibilityReason(ctx, value, platform, model, compact, required); reason != "" {
		return reason
	}
	if required, _ := ctx.Value(imageModelRequiredKey{}).(bool); required && !media.IsImageGenerationModel(provider.ExecutionModelPolicy(value).UpstreamModel(ctx, model)) {
		return "image_model_required"
	}
	if preferred, configured := ctx.Value(preferredAccountsKey{}).(map[int64]struct{}); configured {
		if _, found := preferred[value.Record.ID]; !found {
			return "model_routing_deferred"
		}
	}
	state, _ := ctx.Value(candidatePolicyKey{}).(candidatePolicy)
	if state.groupID != nil && !openAIStickyAccountMatchesGroup(value, state.groupID) {
		return "group_mismatch"
	}
	group, _ := requeststate.GroupFromContext(ctx)
	if group != nil && group.RequireOAuthOnly && !value.View().IsOAuth() {
		return "oauth_required"
	}
	if s.generic != nil {
		if !s.generic.isAccountSchedulableForQuota(value) {
			return "quota_exceeded"
		}
		if !s.generic.isAccountSchedulableForWindowCost(ctx, value, value.Record.ID == state.stickyID) {
			return "window_cost_exceeded"
		}
		if !s.generic.isAccountSchedulableForRPM(ctx, value, value.Record.ID == state.stickyID) {
			return "rpm_exceeded"
		}
		if s.generic.isAccountBlockedBySchedulingThreshold(ctx, value) {
			return "scheduling_threshold"
		}
	}
	if s.gemini != nil && value.Record.Platform == capability.PlatformGemini && !s.gemini.passesRateLimitPreCheckWithCache(ctx, value, model, nil) {
		return "gemini_quota_exhausted"
	}
	return ""
}

type (
	selectionRequestKey struct{}
	selectionRequest    struct {
		groupID *int64
		model   string
	}
)

// withSelectionRequest 固定模型和分组，补全凭据时可用最新账号再次核对资格。
func withSelectionRequest(ctx context.Context, groupID *int64, model string) context.Context {
	return context.WithValue(ctx, selectionRequestKey{}, selectionRequest{groupID: groupID, model: model})
}
