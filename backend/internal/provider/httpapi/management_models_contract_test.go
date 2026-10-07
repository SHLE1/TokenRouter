package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	catalogtest "github.com/TokenFlux/TokenRouter/internal/modelcatalog/testkit"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"

	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient/tlsfingerprint"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	routingprovider "github.com/TokenFlux/TokenRouter/internal/routing/provider"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type availableModelsAdminService struct {
	*managementMutationFixture
	provider providercore.Record
}

func (s *availableModelsAdminService) GetProvider(_ context.Context, id int64) (*providercore.Record, error) {
	if s.provider.ID == id {
		acc := s.provider
		return &acc, nil
	}
	return s.managementMutationFixture.GetProvider(context.Background(), id)
}

func setupAvailableModelsRouter(adminSvc ProviderManagement) *gin.Engine {
	router := gin.New()
	handler := NewManagementHandler(adminSvc, ManagementOptions{Catalog: routing.NewAdminCatalog(routingprovider.AdminCatalogOptions(catalogtest.New("catalog-model", "gpt-5.4", "gemini-new-model", "glm-4.7"))), ModelDefaults: provideradapter.ModelDefaults(), ModelSupports: func(_ context.Context, v *providercore.Record, id string) bool {
		return v.IsModelSupported(id, provideradapter.ModelDefaults(), provideradapter.ModelRules(v))
	}})
	router.GET("/api/v1/admin/providers/:id/models", handler.GetAvailableModels)
	return router
}

type syncUpstreamHTTPUpstream struct {
	resp *http.Response
	err  error
}

func (u *syncUpstreamHTTPUpstream) Do(req *http.Request, proxyURL string, providerID int64, providerConcurrency int) (*http.Response, error) {
	if u.err != nil {
		return nil, u.err
	}
	return u.resp, nil
}

func (u *syncUpstreamHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, providerID int64, providerConcurrency int, profile *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, providerID, providerConcurrency)
}

func setupSyncUpstreamModelsRouter(adminSvc ProviderManagement, upstream provideradapter.QoderTransport) *gin.Engine {
	router := gin.New()
	modelPolicy := egress.OperatorURLPolicy{}
	catalogue := &provideradapter.ModelCatalogue{Transport: upstream, Options: provideradapter.ModelCatalogueOptions{ValidateURL: modelPolicy.Validate, OperatorValidator: modelPolicy.Validate, BodyLimit: 8 * 1024 * 1024, CodexModelsURL: provideradapter.DefaultCodexModelsURL}}
	models := providercore.NewModelSyncService(catalogue.FetchUpstreamSupportedModels)
	handler := NewManagementHandler(adminSvc, ManagementOptions{Models: models})
	router.POST("/api/v1/admin/providers/:id/models/sync-upstream", handler.SyncUpstreamModels)
	router.POST("/api/v1/admin/providers/models/sync-upstream-preview", handler.SyncUpstreamModelsPreview)
	return router
}

// TestAvailableModelsUsesUnifiedCatalog 覆盖认证、白名单、映射和国产平台的统一候选。
func TestAvailableModelsUsesUnifiedCatalog(t *testing.T) {
	parent := int64(1)
	for _, tc := range []struct {
		name, platform, kind string
		credentials          map[string]any
		extra                map[string]any
		parent               *int64
		includes, excludes   []string
	}{
		{name: "empty", platform: "anthropic", kind: "apikey", includes: []string{"catalog-model", "gpt-5.4"}},
		{name: "oauth", platform: "openai", kind: "oauth", includes: []string{"catalog-model", "gpt-5.4"}, excludes: []string{"gemini-new-model", "glm-4.7"}},
		{name: "explicit", platform: "openai", kind: "oauth", credentials: map[string]any{"model_whitelist": []string{"custom"}}, includes: []string{"custom"}, excludes: []string{"catalog-model", "gpt-5.4"}},
		{name: "passthrough", platform: "openai", kind: "oauth", extra: map[string]any{"openai_passthrough": true}, credentials: map[string]any{"model_whitelist": []string{"custom"}}, includes: []string{"custom"}, excludes: []string{"gpt-5.4"}},
		{name: "mapping", platform: "openai", kind: "apikey", credentials: map[string]any{"model_mapping": map[string]any{"alias": "unknown-target"}}, includes: []string{"alias", "unknown-target", "catalog-model"}},
		{name: "empty whitelist", platform: "openai", kind: "apikey", credentials: map[string]any{"model_whitelist": []string{}, "model_mapping": map[string]any{"alias": "alias"}}, includes: []string{"alias", "catalog-model"}},
		{name: "wildcard", platform: "openai", kind: "apikey", credentials: map[string]any{"model_whitelist": []string{"gpt-*"}}, includes: []string{"gpt-5.4"}, excludes: []string{"catalog-model"}},
		{name: "spark", platform: "openai", kind: "oauth", parent: &parent, includes: []string{"gpt-5.3-codex-spark"}, excludes: []string{"gpt-5.4", "catalog-model"}},
		{name: "google one", platform: "gemini", kind: "oauth", credentials: map[string]any{"oauth_type": "google_one"}, includes: []string{"gemini-new-model", "catalog-model"}},
		{name: "qoder cn", platform: "qoder", kind: "cosy", credentials: map[string]any{"site": "cn"}, includes: []string{"qwen3.6-flash", "catalog-model"}, excludes: []string{"claude-opus-4-6"}},
		{name: "qoder global", platform: "qoder", kind: "cosy", credentials: map[string]any{"site": "global"}, includes: []string{"claude-opus-4-6", "catalog-model"}, excludes: []string{"qwen3.6-flash"}},
		{name: "kimi", platform: "kimi", kind: "apikey", includes: []string{"catalog-model"}},
		{name: "zhipu", platform: "zhipu", kind: "apikey", includes: []string{"glm-4.7"}},
		{name: "deepseek", platform: "deepseek", kind: "apikey", includes: []string{"catalog-model"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &availableModelsAdminService{managementMutationFixture: newManagementMutationFixture(), provider: providercore.Record{ID: 42, Platform: tc.platform, Type: tc.kind, Credentials: tc.credentials, Extra: tc.extra, ParentProviderID: tc.parent}}
			rec := httptest.NewRecorder()
			setupAvailableModelsRouter(svc).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/providers/42/models", nil))
			require.Equal(t, http.StatusOK, rec.Code)
			var response struct {
				Data []struct {
					ID   string `json:"id"`
					Name string `json:"display_name"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			ids := []string{}
			for _, model := range response.Data {
				ids = append(ids, model.ID)
				require.NotEmpty(t, model.Name)
			}
			for _, id := range tc.includes {
				require.Contains(t, ids, id)
			}
			for _, id := range tc.excludes {
				require.NotContains(t, ids, id)
			}
		})
	}
}

func TestProviderHandlerSyncUpstreamModels_ConfigErrorReturnsBadRequest(t *testing.T) {
	svc := &availableModelsAdminService{
		managementMutationFixture: newManagementMutationFixture(),
		provider: providercore.Record{
			ID:       44,
			Name:     "openai-apikey-missing-key",
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Status:   billing.StatusActive,
			Credentials: map[string]any{
				"base_url": "https://openai.example.com/v1",
			},
		},
	}
	router := setupSyncUpstreamModelsRouter(svc, &syncUpstreamHTTPUpstream{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/providers/44/models/sync-upstream", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "No OpenAI API key is available")
}

func TestProviderHandlerSyncUpstreamModels_UpstreamErrorDoesNotExposeBody(t *testing.T) {
	svc := &availableModelsAdminService{
		managementMutationFixture: newManagementMutationFixture(),
		provider: providercore.Record{
			ID:       45,
			Name:     "openai-apikey-upstream-error",
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Status:   billing.StatusActive,
			Credentials: map[string]any{
				"api_key":  "openai-key",
				"base_url": "https://openai.example.com/v1",
			},
		},
	}
	upstream := &syncUpstreamHTTPUpstream{resp: &http.Response{
		StatusCode: http.StatusBadGateway,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":"SECRET_TOKEN should not be exposed"}`)),
	}}
	router := setupSyncUpstreamModelsRouter(svc, upstream)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/providers/45/models/sync-upstream", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
	require.Contains(t, rec.Body.String(), "Upstream model list request failed with HTTP 502")
	require.NotContains(t, rec.Body.String(), "SECRET_TOKEN")
}

func TestProviderHandlerSyncUpstreamModelsPreview_UsesProvidedCredentials(t *testing.T) {
	router := setupSyncUpstreamModelsRouter(newManagementMutationFixture(), &syncUpstreamHTTPUpstream{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"data":[{"id":"gpt-5.1"},{"id":"gpt-5.1"},{"id":"o3"}]}`)),
	}})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/providers/models/sync-upstream-preview",
		strings.NewReader(`{"platform":"openai","type":"apikey","base_url":"https://openai.example.com/v1","api_key":"openai-key"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data struct {
			Models []string `json:"models"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, []string{"gpt-5.1", "o3"}, resp.Data.Models)
}

func TestProviderHandlerSyncUpstreamModelsPreview_ConfigErrorReturnsBadRequest(t *testing.T) {
	router := setupSyncUpstreamModelsRouter(newManagementMutationFixture(), &syncUpstreamHTTPUpstream{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/providers/models/sync-upstream-preview",
		strings.NewReader(`{"platform":"openai","type":"apikey","base_url":"https://openai.example.com/v1"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "required")
}
