package provider

import (
	"maps"
	"slices"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

// snapshot.go builds provider identity and protocol snapshots for candidate
// routing checks.
// Candidate protocol checks need provider identity and eligibility while
// credentials stay in the execution record. RoutingSnapshot captures those
// fields, resolves configured protocols, and detaches its protocol slice and
// optional identity/expiry pointers for later candidate checks. The diagnostic
// response-header name is copied for execution observation before dispatch.

// ProviderSnapshot holds provider identity, eligibility, and runtime selection data.
type ProviderSnapshot struct {
	// ModelPolicy captures OpenCode aliases before protocol resolution. Other
	// callers populate it separately before CandidatePlan.ResolveModel applies
	// their configured provider mappings.
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

// RoutingSnapshot copies selection fields with protocols resolved from the
// record's configuration. Nil records return a zero snapshot; parent, expiry,
// and protocol values are detached. OpenCode records also capture their
// aliases for model-specific protocol matching.
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
	if r.IsOpenCode() {
		snapshot.ModelPolicy = NewModelRoutingSnapshot(ResolveModelMapping(r, ModelMappingDefaults{}))
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

// Protocols returns provider capabilities with copied enabled protocols and
// model encodings. Callers can change both collections independently; nil and
// empty enabled selections retain their meaning during capability resolution.
func (s ProviderSnapshot) Protocols() capability.ProviderProtocols {
	return capability.ProviderProtocols{Platform: s.Platform, Type: s.Type, AuthMode: s.AuthMode, Enabled: slices.Clone(s.EnabledProtocols), ModelProtocols: maps.Clone(s.ModelProtocols)}
}
