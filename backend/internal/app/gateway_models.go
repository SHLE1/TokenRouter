package app

import (
	"context"

	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/googleforward"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/selection"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/upstream/gemini"
	"github.com/gin-gonic/gin"
)

// provideModelsHTTP 构造四个模型目录查询入口。
func provideModelsHTTP(catalogue *routing.RequestableCatalogue, reader *googleforward.Gemini, activity *gatewayRequestActivity, choices *selection.Gemini) *gatewayhttp.ModelsHandler {
	ports := gatewayhttp.ModelsPorts{
		ReadAccess: keyhttp.GetAPIKeyFromContext, ReadPlatform: keyhttp.GetForcePlatformFromContext,
		ReadBilling: func(c *gin.Context) (*billing.APIKeyBillingContext, bool) {
			return gatewayhttp.GetAPIKeyBillingContext(c)
		},
		SafeSegment: gemini.IsSafeGeminiModelPathSegment,
		SelectModel: func(ctx context.Context, id *int64) (gatewayhttp.GeminiModelReader, error) {
			value, err := choices.SelectProviderForAIStudioEndpoints(ctx, id)
			if err != nil {
				return nil, err
			}
			return geminiModelReadTarget{reader, value}, nil
		},
		CheckAntigravity: choices.HasAntigravityProviders,
	}
	if catalogue != nil {
		ports.Catalogue = catalogue
	}
	result := gatewayhttp.NewModelsHandler(ports, gatewayprovider.ModelDisplayCatalogue{})
	if activity != nil {
		result.BindRequestActivity(activity.Enter)
	}
	return result
}

// geminiModelReadTarget 保存本次选中的提供商，供模型资源查询使用。
type geminiModelReadTarget struct {
	reader *googleforward.Gemini
	value  *gatewayprovider.ExecutionProvider
}

func (p geminiModelReadTarget) Read(ctx context.Context, path string) (*gatewayhttp.ModelHTTPResponse, error) {
	value, err := p.reader.ForwardAIStudioGET(ctx, p.value, path)
	if value == nil {
		return nil, err
	}
	return &gatewayhttp.ModelHTTPResponse{StatusCode: value.StatusCode, Headers: value.Headers, Body: value.Body}, err
}
