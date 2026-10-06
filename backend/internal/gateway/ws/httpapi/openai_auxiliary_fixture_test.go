package httpapi

import (
	"context"
	"time"

	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressprovider "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// auxiliaryFixtureInputs 提供辅助请求测试所需的依赖。
type auxiliaryFixtureInputs struct {
	allowHTTP     bool
	transport     httpclient.UpstreamTransport
	profiles      *egressprovider.TLSProfiles
	store         gatewayadapter.ExecutionProviderStore
	credentials   *provider.OpenAIExecutionCredentials
	observer      *provideradapter.UpstreamHealth
	authorization *provider.OpenAIAuthorization
}

func newAuxiliaryFixture(v auxiliaryFixtureInputs) *gatewayhttp.OpenAIAuxiliary {
	blocks := provider.NewRuntimeBlockState(time.Now)
	models := provider.NewModelTransientState(0)
	credentials := v.credentials
	if credentials == nil {
		credentials = &provider.OpenAIExecutionCredentials{}
	}
	if v.store != nil {
		credentials.Parent = func(ctx context.Context, id int64) (*provider.Record, error) {
			a, err := v.store.GetByID(ctx, id)
			return gatewayadapter.ExecutionRecord(a), err
		}
	}
	identity := gatewayadapter.NewExecutionAgentIdentity(&provider.OpenAITaskCoordinator{}, v.store, nil, nil)
	turns := &gatewayhttp.CodexTurnStateHeaders{Origins: session.NewCodexTurnOrigins(time.Now), TTL: func() time.Duration { return time.Hour }}
	requests := &gatewayhttp.OpenAIRequests{Options: gatewayhttp.OpenAIRequestOptions{URLPolicy: egress.OperatorURLPolicy{AllowInsecureHTTP: v.allowHTTP}}, Providers: v.store, Identity: identity, Credentials: credentials, Transport: v.transport, Profiles: v.profiles, Turns: turns, ClientPolicy: &provideradapter.OpenAIProbePolicy{Available: true, DefaultBrowserUserAgent: gateway.DefaultOpenAICodexUserAgent, Profiles: v.profiles}, Failure: &gatewayhttp.UpstreamTransportFailure{Health: &provideradapter.TransportHealth{Runtime: blocks}}}
	grok := &provideradapter.GrokHealth{Store: v.store, Health: v.observer, Runtime: blocks, ModelTransient: models, NormalizeModel: func(value *provider.Record, model string) string {
		return (gatewayadapter.ModelPolicy{Record: value}).NormalizeOpenAI(model)
	}}
	output := &gatewayhttp.OpenAIResponseOutput{Options: gatewayhttp.OpenAIResponseOptions{Configured: true, ReadLimit: 128 * 1024 * 1024}, Health: &provideradapter.OpenAIResponseHealth{Health: v.observer, Runtime: blocks, ModelTransient: models}, GrokHealth: grok, Headers: egress.CompileHeaderFilter(egress.ResponseHeaderOptions{})}
	return &gatewayhttp.OpenAIAuxiliary{Requests: requests, Output: output, Authorization: v.authorization, CodexUsage: &provideradapter.CodexUsageObserver{Store: v.store, Throttle: provider.NewWriteThrottle(30 * time.Second)}}
}
