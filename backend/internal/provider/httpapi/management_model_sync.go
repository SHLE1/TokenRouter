package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	response "github.com/TokenFlux/TokenRouter/internal/server/httpx"
)

// management_model_sync.go handles live model queries for saved providers and
// unsaved credential previews.
// Live model queries use the model-sync use case for stored or temporary
// credentials. The HTTP layer distinguishes local configuration failures from
// upstream failures and sends sanitized messages; preview constructs an unsaved
// record so query results can be inspected before provider creation.

// SyncUpstreamModels loads provider credentials and requests a live
// supported-model list. OpenCode accepts temporary mode and root settings on
// a copied record, and returns model protocol suggestions alongside IDs.
// Lookup failures map to 404, configuration or
// unsupported errors to 400, and other query failures to sanitized 502
// responses.
// POST /api/v1/admin/providers/:id/models/sync-upstream
func (h *ManagementHandler) SyncUpstreamModels(c *gin.Context) {
	providerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid provider ID")
		return
	}

	provider, err := h.adminService.GetProvider(c.Request.Context(), providerID)
	if err != nil {
		response.NotFound(c, "Provider not found")
		return
	}

	if h.models == nil {
		response.InternalError(c, "Provider test service is not configured")
		return
	}

	if provider.IsOpenCode() && c.Request.ContentLength != 0 {
		var override ProviderModelsSyncRequest
		if err := response.BindJSONStrict(c, &override); err != nil {
			response.BadRequest(c, "Invalid model synchronization settings")
			return
		}
		provider = providercore.CloneRecord(provider)
		if override.ProviderMode != "" {
			provider.Credentials["provider_mode"] = override.ProviderMode
		}
		if override.BaseURL != nil {
			provider.Credentials["base_url"] = *override.BaseURL
		}
	}

	models, err := h.models.Fetch(c.Request.Context(), provider)
	if err != nil {
		if syncErr, ok := errors.AsType[*providercore.UpstreamModelSyncError](err); ok {
			switch syncErr.Kind {
			case providercore.UpstreamModelSyncErrorConfiguration, providercore.UpstreamModelSyncErrorUnsupported:
				response.BadRequest(c, syncErr.SafeMessage())
			default:
				slog.Warn("sync_upstream_models_failed", "provider_id", providerID, "kind", syncErr.Kind)
				response.Error(c, http.StatusBadGateway, syncErr.SafeMessage())
			}
			return
		}

		slog.Warn("sync_upstream_models_failed", "provider_id", providerID)
		response.Error(c, http.StatusBadGateway, "Failed to sync upstream models from upstream")
		return
	}

	result := gin.H{"models": models}
	if protocols := h.models.Protocols(provider, models); protocols != nil {
		result["model_protocols"] = protocols
	}
	response.Success(c, result)
}

// SyncUpstreamModelsPreview strictly binds temporary API-key credentials and
// queries an unsaved provider record. It applies the same sanitized error
// categories as stored-provider synchronization, and the handler performs no
// provider configuration write.
// POST /api/v1/admin/providers/models/sync-upstream-preview
func (h *ManagementHandler) SyncUpstreamModelsPreview(c *gin.Context) {
	var req ProviderModelsSyncPreviewRequest
	if err := response.BindJSONStrict(c, &req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	tempProvider := &providercore.Record{
		Platform: req.Platform,
		Type:     req.Type,
		Credentials: map[string]any{
			"api_key":  req.APIKey,
			"base_url": req.BaseURL,
		},
	}

	if req.ProviderMode != "" {
		tempProvider.Credentials["provider_mode"] = req.ProviderMode
	}

	if h.models == nil {
		response.InternalError(c, "Provider test service is not configured")
		return
	}

	models, err := h.models.Fetch(c.Request.Context(), tempProvider)
	if err != nil {
		if syncErr, ok := errors.AsType[*providercore.UpstreamModelSyncError](err); ok {
			switch syncErr.Kind {
			case providercore.UpstreamModelSyncErrorConfiguration, providercore.UpstreamModelSyncErrorUnsupported:
				response.BadRequest(c, syncErr.SafeMessage())
			default:
				slog.Warn("sync_upstream_models_preview_failed", "platform", req.Platform, "kind", syncErr.Kind)
				response.Error(c, http.StatusBadGateway, syncErr.SafeMessage())
			}
			return
		}

		slog.Warn("sync_upstream_models_preview_failed", "platform", req.Platform)
		response.Error(c, http.StatusBadGateway, "Failed to sync upstream models from upstream")
		return
	}

	result := gin.H{"models": models}
	if protocols := h.models.Protocols(tempProvider, models); protocols != nil {
		result["model_protocols"] = protocols
	}
	response.Success(c, result)
}
