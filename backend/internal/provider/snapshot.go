package provider

import (
	"slices"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/modelcatalog"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// snapshot.go captures provider routing data and resolves one model attempt.
// Candidate checks need provider identity and eligibility while credentials
// stay in the execution record. RoutingSnapshot copies configured aliases,
// enabled protocols, optional identity pointers, and OpenCode offering metadata
// before a route is selected. Adapter callers add their platform mapping
// defaults to that captured configuration.
//
// Model resolution performs one alias hop and selects an encoding from the
// resulting identity. OpenCode offerings can assign different protocols to the
// same name in payg and Go, so the selected model, protocol, and canonical
// attribute identity travel together. Execution consumers retain this result;
// retries resolve a fresh snapshot, and structural capability checks use the
// separate Protocols view when no model has been chosen.

// ProviderSnapshot holds provider identity, eligibility, and runtime selection data.
type ProviderSnapshot struct {
	// ModelPolicy captures configured aliases before protocol resolution. Adapter
	// callers replace it with platform defaults when their execution policy supplies them.
	ModelPolicy ModelRoutingSnapshot `json:"-"`
	ID          int64                // provider record identifier
	// ParentProviderID retains optional parent identity. RoutingSnapshot copies
	// its value so editing this pointer does not change the source record's parent.
	ParentProviderID *int64 // optional credential parent identifier
	Platform         string // configured upstream platform identity
	Type             string // configured provider authentication type
	AuthMode         string // resolved protocol authentication mode
	// ModelProtocols holds this snapshot's independently owned OpenCode encoding
	// table. Other platforms leave it nil because their native choices are model independent.
	ModelProtocols map[string]capability.ProtocolID // exact final-model protocol choices for OpenCode
	// ModelOfferings retains mode-specific canonical identities from the same
	// publication as protocol defaults. Protocol overrides leave these identities intact.
	ModelOfferings map[string]modelcatalog.OpenCodeOffering
	// EnabledProtocols is resolved and copied by RoutingSnapshot; Protocols
	// copies it again before returning a capability view to its caller.
	EnabledProtocols []capability.ProtocolID // resolved upstream protocol choices
	Status           string                  // stored provider lifecycle status
	Schedulable      bool                    // configured candidate scheduling flag
	Concurrency      int                     // configured request concurrency limit
	RequestIDHeader  string                  `json:"-"` // selected diagnostic response header
	Priority         int                     // configured candidate ordering priority
	// ExpiresAt retains the optional stored expiry. RoutingSnapshot copies that
	// value so changing the snapshot's timestamp leaves the source record unchanged.
	ExpiresAt *time.Time // optional provider expiry time
}

// ResolvedModelRoute binds the outgoing model and encoding selected from one snapshot.
// Consumers send Model directly and resolve a fresh result after selecting a
// different provider or configuration, including retries that change service mode.
type ResolvedModelRoute struct {
	Model string // final upstream model identity
	// AttributeModel is empty when bundled offerings have no canonical reference;
	// attribute readers then consult their retained live catalog for that offering.
	AttributeModel string                // canonical author metadata identity
	Protocol       capability.ProtocolID // selected upstream execution protocol
	MappingMatched bool                  // provider alias match indicator
}

// RoutingSnapshot copies selection fields with protocols resolved from the
// record's configuration. Nil records return a zero snapshot; parent, expiry,
// and protocol values are detached. Every platform captures configured aliases;
// OpenCode also captures its mode-specific canonical metadata identities.
// Keep the source record stable during capture so its identity and protocol
// fields describe the same configuration.
func (r *Record) RoutingSnapshot() ProviderSnapshot {
	if r == nil {
		return ProviderSnapshot{}
	}
	snapshot := ProviderSnapshot{
		ID: r.ID, Platform: r.Platform, Type: r.Type, AuthMode: ProtocolAuthMode(r),
		ModelProtocols:   r.OpenCodeModelProtocols(),
		EnabledProtocols: slices.Clone(r.UpstreamProtocols()), Status: r.Status, Schedulable: r.Schedulable,
		Concurrency: r.Concurrency, Priority: r.Priority, RequestIDHeader: UpstreamRequestIDHeaderName(r),
	}
	// Core snapshots capture saved mappings; execution adapters install defaults
	// for platforms such as Antigravity before resolving an executable attempt.
	snapshot.ModelPolicy = NewModelRoutingSnapshot(ResolveModelMapping(r, ModelMappingDefaults{Antigravity: func() map[string]string { return nil }}))
	if r.IsOpenCode() {
		snapshot.ModelOfferings = modelcatalog.OpenCodeOfferings(r.OpenCodeMode())
	}
	if r.ParentProviderID != nil {
		id := *r.ParentProviderID
		snapshot.ParentProviderID = &id
	}
	if r.ExpiresAt != nil {
		at := *r.ExpiresAt
		snapshot.ExpiresAt = &at
	}
	return snapshot
}

// Protocols returns structural capabilities with copied enabled protocols.
// Callers use ProtocolsForModel for executable model checks. Nil and empty
// enabled selections retain their meaning during capability resolution.
func (s ProviderSnapshot) Protocols() capability.ProviderProtocols {
	return capability.ProviderProtocols{Platform: s.Platform, Type: s.Type, AuthMode: s.AuthMode, Enabled: slices.Clone(s.EnabledProtocols)}
}

// ProtocolsForModel returns capabilities narrowed to an already mapped model.
// OpenCode exposes its enabled declared encoding, returning an empty selection
// for unknown or disabled models. Callers must pass the final one-hop identity.
func (s ProviderSnapshot) ProtocolsForModel(model string) capability.ProviderProtocols {
	result := s.Protocols()
	if s.Platform == PlatformOpenCode {
		target, known := s.ModelProtocols[model]
		result.Enabled = []capability.ProtocolID{}
		if known && slices.Contains(s.EnabledProtocols, target) {
			result.Enabled = []capability.ProtocolID{target}
		}
	}
	return result
}

// ResolveModelRoute selects the final model, metadata identity, and protocol
// from this snapshot. Mapping executes once, and a missing source selects the
// first enabled native protocol for automatic model probes.
//
// OpenCode narrows native choices to the model's declared encoding before
// applying client conversion policy. Unknown models and explicitly disabled
// protocols fail even when the provider supports the protocol family generally.
// Callers keep the returned values together during an attempt and resolve
// again against a fresh snapshot when they retry another configuration.
func (s ProviderSnapshot) ResolveModelRoute(requested string, source capability.ProtocolID, fallbacks map[capability.ProtocolID][]capability.ProtocolID) (ResolvedModelRoute, bool) {
	model, matched := s.ModelPolicy.Resolve(requested)
	protocols := s.ProtocolsForModel(model)
	if source == "" {
		for _, target := range protocols.Enabled {
			if slices.Contains(capability.NativeProtocolOptions(s.Platform, s.Type, s.AuthMode), target) {
				source = target
				break
			}
		}
	}
	target, ok := capability.ResolveRoute(protocols, source, fallbacks)
	if !ok {
		return ResolvedModelRoute{}, false
	}
	return ResolvedModelRoute{Model: model, AttributeModel: s.ModelOfferings[model].CanonicalModelID, Protocol: target, MappingMatched: matched}, true
}
