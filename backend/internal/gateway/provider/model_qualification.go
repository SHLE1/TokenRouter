package provider

import (
	"context"
	"slices"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/TokenFlux/TokenRouter/internal/protocol"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// model_qualification.go checks protocol and model-window eligibility and
// reports remaining cooldown time.
// A provider can have a model-specific cooldown and a shared-scope cooldown
// applying to the same request. Qualification checks protocol permission and
// provider health before asking whether those scopes permit execution;
// reporting their remaining wait uses the longest active duration because
// the expiry windows run concurrently.
//
// Candidate snapshots copy enabled protocols for route resolution, while the
// current group determines OAuth, privacy, and protocol requirements. Limit
// inspection remains separate from permission: Antigravity overages can admit
// a model with an active cooldown, so Limited reports raw window activity and
// LimitRemaining reports its observed wait even when Schedulable allows work.

// CandidateSnapshot returns routing provider data with copied enabled
// protocols resolved for the current attempt. Nil records produce an empty
// snapshot; protocol selection applies the target's legacy API protocol when
// needed.
func (p ModelPolicy) CandidateSnapshot() provider.ProviderSnapshot {
	if p.Record == nil {
		return provider.ProviderSnapshot{}
	}
	snapshot := p.Record.RoutingSnapshot()
	mapping := provider.ResolveModelMapping(p.Record, provideradapter.ModelDefaults())
	if p.models != nil {
		mapping = p.models.Mapping()
	}
	snapshot.ModelPolicy = provider.NewModelRoutingSnapshot(mapping)
	snapshot.EnabledProtocols = slices.Clone(p.Record.UpstreamProtocolsForLegacy(p.protocolTarget().GetAPIProtocol()))
	return snapshot
}

// ProtocolRoute checks provider protocol support against the supplied group
// and client protocol independently of model identity. Nil records return false;
// execution callers subsequently resolve a concrete model with AllowsProtocol.
func (p ModelPolicy) ProtocolRoute(group *routing.Group, source protocol.ProtocolID) (protocol.ProtocolID, bool) {
	if p.Record == nil {
		return "", false
	}
	var fallbacks map[capability.ProtocolID][]capability.ProtocolID
	if group != nil {
		fallbacks = group.ProtocolFallbacks
	}
	return capability.ResolveRoute(p.CandidateSnapshot().Protocols(), source, fallbacks)
}

// HasProtocolRoute checks structural protocol support and group requirements.
// It serves preliminary pool checks that have no model; executable requests
// must additionally call AllowsProtocol with their requested model identity.
func (p ModelPolicy) HasProtocolRoute(ctx context.Context) bool {
	group, _ := requeststate.GroupFromContext(ctx)
	if p.Record == nil || group != nil && (group.RequireOAuthOnly && !p.Record.IsOAuth() || group.RequirePrivacySet && !p.Record.IsPrivacySet()) {
		return false
	}
	source, _ := requeststate.ClientProtocolFromContext(ctx)
	if source != "" {
		_, ok := p.ProtocolRoute(group, source)
		return ok
	}
	// Model-independent callers also use this method for health inspection
	// without a client entry; their eventual execution path supplies the protocol.
	if !p.Record.IsOpenCode() {
		return true
	}
	snapshot := p.CandidateSnapshot()
	for _, target := range snapshot.EnabledProtocols {
		if slices.Contains(capability.NativeProtocolOptions(snapshot.Platform, snapshot.Type, snapshot.AuthMode), target) {
			return true
		}
	}
	return false
}

// AllowsProtocol resolves an executable route for the supplied requested model.
// Group OAuth and privacy requirements apply first. OpenCode requests without
// a client protocol select an enabled native encoding; other platforms retain
// their protocol-independent health checks when the client entry is absent.
//
// A retained request plan supplies group mapping and conversion policy when its
// model matches this request. Auxiliary model queries build their own route so
// an attempt for another model cannot supply its protocol. Every successful
// OpenCode decision requires the final mapped model's declared protocol to be
// enabled on the captured provider configuration.
func (p ModelPolicy) AllowsProtocol(ctx context.Context, model string) bool {
	if !p.HasProtocolRoute(ctx) {
		return false
	}
	group, _ := requeststate.GroupFromContext(ctx)
	source, _ := requeststate.ClientProtocolFromContext(ctx)
	// Health-only model checks on other platforms historically omit the client
	// protocol. Their execution adapters validate the selected entry when one exists.
	if source == "" && !p.Record.IsOpenCode() {
		return true
	}
	plan := routing.Plan(routing.PlanInput{Group: group, ClientProtocol: source, RequestedModel: model})
	if retained, planned := requeststate.RoutePlanFromContext(ctx); planned && (group == nil || retained.GroupID() == group.ID) {
		models := retained.Models()
		if model == models.RequestedModel || model == models.GroupMappedModel {
			plan = retained.WithClientProtocol(source)
		}
	}
	_, ok := plan.ResolveCandidate(p.CandidateSnapshot())
	return ok
}

// Schedulable applies protocol eligibility, provider scheduling state, and
// applicable model window permission in that order. Nil records fail
// immediately, and each rejected stage prevents later eligibility work.
func (p ModelPolicy) Schedulable(ctx context.Context, model string) bool {
	if p.Record == nil {
		return false
	}
	if !p.AllowsProtocol(ctx, model) || !p.Record.IsSchedulable() {
		return false
	}
	return p.AllowsModel(ctx, model)
}

// Limited reports whether any model or shared-scope limit selected by
// LimitKeys is active. An empty key set yields false, keeping a missing model
// window distinct from an active one.
func (p ModelPolicy) Limited(ctx context.Context, model string) bool {
	return slices.ContainsFunc(p.LimitKeys(ctx, model), p.Record.ModelRateLimitActive)
}

// LimitRemaining returns the longest observed remaining duration across the
// model and shared-scope windows selected by LimitKeys. Empty, invalid, or
// elapsed windows leave zero, using the provider record's clock for each read.
//
// Several scopes can restrict one request simultaneously. Their waits overlap,
// so adding durations would overstate the cooldown; the maximum identifies the
// last observed expiry. This is raw window evidence, and callers use
// Schedulable separately when overage policy can permit execution during a
// model cooldown.
func (p ModelPolicy) LimitRemaining(ctx context.Context, model string) time.Duration {
	remaining := time.Duration(0)
	for _, key := range p.LimitKeys(ctx, model) {
		if value := p.Record.ModelRateLimitRemaining(key); value > remaining {
			remaining = value
		}
	}
	return remaining
}

// FinalAntigravityModel resolves the upstream model using the request's
// optional thinking hint. It delegates directly to the provider adapter, which
// owns platform mapping and unsupported-model handling.
func (p ModelPolicy) FinalAntigravityModel(ctx context.Context, model string) string {
	return provideradapter.FinalAntigravityModel(p.Record, model, modelThinking(ctx))
}
