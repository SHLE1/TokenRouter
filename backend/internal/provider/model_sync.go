package provider

import (
	"context"
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/modelcatalog"
)

// model_sync.go fetches live model catalogues using copied provider records
// and tracked request lifetimes.
// Administrator model previews fetch a live catalogue using a provider snapshot
// whose credentials can be mutated by the adapter. The service copies that
// input before fetching and copies the returned model slice before handing it
// back to its caller.
//
// The shared operationActivity implementation tracks these request-triggered
// queries, combines request and shutdown cancellation, and fixes the first stop
// result. Construction performs no upstream query; requests begin work when
// Query is called, and stopped services reject another query.

// ModelSyncService tracks request-triggered model queries through an injected
// fetch function.
type ModelSyncService struct {
	// fetch receives the service-owned provider copy for one registered query.
	// It observes cancellation so shutdown can wait for the call to return.
	fetch func(context.Context, *Record) ([]string, error)
	// activity owns admission, cancellation, and the retained shutdown result.
	// Each successful registration remains active until Query's deferred finish.
	activity operationActivity
	// CatalogProtocols reads mode-specific encodings from the shared catalog.
	// A missing callback uses the bundled publication for local adapters and tests.
	CatalogProtocols func(string) map[string]string
}

// ModelSyncResult contains live identifiers and independently owned catalogue suggestions.
type ModelSyncResult struct {
	Models []string `json:"models"` // live upstream model identifiers
	// ModelProtocols contains catalogue suggestions for the returned models. Saved
	// administrator overrides remain owned by the provider configuration and form.
	ModelProtocols map[string]string `json:"model_protocols,omitempty"`
}

// ModelSyncPreview supplies caller-owned credentials for an unsaved model query.
type ModelSyncPreview struct {
	Platform string // temporary upstream platform
	Type     string // temporary authentication kind
	APIKey   string // caller supplied credential
	BaseURL  string // temporary service root
	// ProviderMode selects the OpenCode service; an empty value leaves the temporary
	// record unconfigured so the provider uses its default payg mode.
	ProviderMode string
}

// NewModelSyncService constructs a request-triggered catalogue service around
// fetch. A nil function returns nil to represent an unavailable service, and
// construction performs no upstream request or persistence.
func NewModelSyncService(fetch func(context.Context, *Record) ([]string, error)) *ModelSyncService {
	if fetch == nil {
		return nil
	}
	return &ModelSyncService{fetch: fetch}
}

// Query registers a catalogue query and passes a copied provider record to the
// adapter. Its result pairs live model IDs with pure catalogue suggestions,
// leaving saved manual protocol overrides in the caller-owned configuration.
//
// Request cancellation and service shutdown reach the adapter through the
// registered context. Completion unregisters the operation so StopContext can
// finish waiting, and a stopped service returns ErrRefreshStopped before
// invoking fetch. Callers provide stable input during copying and adapters
// respond to cancellation to permit timely shutdown.
func (s *ModelSyncService) Query(ctx context.Context, v *Record) (ModelSyncResult, error) {
	ctx, finish, err := s.activity.begin(ctx, ErrRefreshStopped)
	if err != nil {
		return ModelSyncResult{}, err
	}
	defer finish()
	snapshot := CloneRecord(v)
	models, err := s.fetch(ctx, snapshot)
	if err != nil {
		return ModelSyncResult{}, err
	}
	return ModelSyncResult{Models: slices.Clone(models), ModelProtocols: s.protocolSuggestions(snapshot, models)}, nil
}

// Preview queries models using the credentials supplied by the caller. It
// constructs an unsaved record inside this use case and shares Query's copy,
// cancellation, and catalogue suggestion rules.
func (s *ModelSyncService) Preview(ctx context.Context, input ModelSyncPreview) (ModelSyncResult, error) {
	credentials := map[string]any{
		"api_key":  input.APIKey,
		"base_url": input.BaseURL,
	}
	if input.ProviderMode != "" {
		credentials["provider_mode"] = input.ProviderMode
	}
	return s.Query(ctx, &Record{Platform: input.Platform, Type: input.Type, Credentials: credentials})
}

// StopContext rejects subsequent model queries, cancels active queries, and
// waits within ctx's budget. The shared activity tracker retains the first stop
// outcome, including a timeout while an adapter remains unfinished.
func (s *ModelSyncService) StopContext(ctx context.Context) error {
	return s.activity.stop(ctx, "provider model list requests")
}

// protocolSuggestions returns catalogue encodings for the queried OpenCode
// model IDs. A published mode table is authoritative, including omitted models;
// absent publication uses bundled defaults. Administrator overrides stay in the
// caller's configuration so deleting a draft override restores catalogue choice.
func (s *ModelSyncService) protocolSuggestions(value *Record, models []string) map[string]string {
	if !value.IsOpenCode() {
		return nil
	}
	// An unavailable live publication retains bundled model knowledge, while a
	// published empty table intentionally supplies no current protocol suggestions.
	routes := modelcatalog.OpenCodeProtocols(value.OpenCodeMode())
	if s.CatalogProtocols != nil {
		if current := s.CatalogProtocols(value.OpenCodeMode()); current != nil {
			routes = current
		}
	}
	result := make(map[string]string)
	for _, model := range models {
		if target := routes[model]; target != "" {
			result[model] = target
		}
	}
	return result
}
