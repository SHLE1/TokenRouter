package httpapi

import (
	"net/http"
	"time"

	gatewayws "github.com/TokenFlux/TokenRouter/internal/gateway/ws"

	openaiws "github.com/TokenFlux/TokenRouter/internal/upstream/openai/ws"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/selection"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
	gatewaytestkit "github.com/TokenFlux/TokenRouter/internal/gateway/testkit"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/settings"
	"github.com/TokenFlux/TokenRouter/internal/upstream/grok"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// wsFixtureOptions 保存传输测试使用的选项。
type wsFixtureOptions struct {
	WS      gatewayws.Parameters
	Pool    openaiws.WSPoolOptions
	Request OpenAIRequestOptions
	Output  OpenAIResponseOptions
}

// wsFixtureInputs 使用实际拥有者与I/O替身，不构造旧网关应用图。
type wsFixtureInputs struct {
	options   *wsFixtureOptions
	providers gatewayadapter.ExecutionProviderStore
	cache     session.GatewayCache
	health    *provideradapter.UpstreamHealth
	transport httpclient.UpstreamTransport
	dialer    openai.WSClientDialer
	pool      *openaiws.WSConnPool
	state     session.OpenAIWSStateStore

	readers   *gatewayadapter.RuntimeReaders
	corrector *openai.CodexToolCorrector
}

// wsExecutionFixture 组合 HTTP 和 WS 测试使用的执行器，共用连接与会话状态。
type wsExecutionFixture struct {
	*OpenAIResponsesExecutor
	Responses *OpenAIResponsesExecutor
	Text      *OpenAITextExecutor
	choices   *selection.Compatible
	options   *wsFixtureOptions
}

func wsFixturePoolOptions(options *wsFixtureOptions) *openaiws.WSPoolOptions {
	if options == nil {
		return nil
	}
	out := options.Pool

	out.DialTimeoutSeconds = options.WS.DialTimeoutSeconds
	out.PrewarmCooldownMS = options.WS.PrewarmCooldownMS
	return &out
}

func newWSFixture(v wsFixtureInputs) *wsExecutionFixture {
	aux := newAuxiliaryFixture(auxiliaryFixtureInputs{transport: v.transport, store: v.providers, observer: v.health})
	requests, output := aux.Requests, aux.Output
	requests.Readers = v.readers
	output.Observer = v.health
	output.Corrector = v.corrector
	output.Headers = nil
	output.Options = OpenAIResponseOptions{ReadLimit: 128 * 1024 * 1024}
	if v.options != nil {
		requests.Options = v.options.Request
		output.Options = v.options.Output
		output.Options.Configured = true
		if output.Options.ReadLimit <= 0 {
			output.Options.ReadLimit = 128 * 1024 * 1024
		}
	}
	state := v.state
	if state == nil {
		state = session.NewOpenAIWSStateStore(v.cache, gatewayadapter.LogOpenAIWSModeInfo)
	}
	output.Responses = state
	output.ProxyCircuit = egress.NewProxyStreamCircuit(egress.DefaultProxyStreamCircuitSettings())
	history, _ := v.cache.(session.ReasoningContentCache)
	output.Reasoning = &session.ReasoningHistory{Cache: history, Warn: gatewayadapter.WarnReasoningCacheFailure}
	output.Redact = requests.Identity.Redact
	choicesOptions := selection.DefaultOptions()
	if v.options != nil {
		ws := v.options.WS
		if ws.StickyResponseIDTTLSeconds > 0 {
			choicesOptions.ResponseTTL = time.Duration(ws.StickyResponseIDTTLSeconds) * time.Second
		}
	}
	choices := selection.NewCompatible(selection.CompatibleDependencies{Reads: selection.Reads{Providers: v.providers}, Shared: selection.Shared{Cache: v.cache, Health: v.health}, Responses: state, RuntimeBlocks: output.Health.Runtime, ModelTransient: output.Health.ModelTransient, ProxyCircuit: output.ProxyCircuit}, choicesOptions)
	output.ResponseTTL = choices.OpenAIHTTPResponseStickyTTL
	requests.Turns.TTL = choices.SessionStickyTTL
	output.Turns = requests.Turns
	if v.readers != nil {
		output.TTFT = v.readers.Gateway.GetOpenAITTFTMode
	}
	credentials := gatewaytestkit.RequestCredentials(v.providers, requests.Credentials, nil, output.Health.Runtime)
	routes := gatewayadapter.GrokRoutes{Validate: grok.ValidateBaseURL}
	if v.options != nil {
		routes.Validate = v.options.Request.URLPolicy.Validate
	}
	if v.readers != nil {
		routes.DefaultMode = v.readers.Gateway.GetGrokDefaultBaseURLMode
	}
	requests.GrokRoutes = routes
	fast := &gatewayadapter.ExecutionFastPolicy{Readers: v.readers}
	connections := openaiws.NewOpenAIWSConnections(wsFixturePoolOptions(v.options), v.dialer, v.pool)
	grokExecutor := &GrokExecutor{Credentials: credentials, Transport: v.transport, Output: output, Health: output.GrokHealth, Routes: routes, Dialer: connections.Dialer(), FastPolicy: fast, Failure: requests.Failure}
	text := &OpenAITextExecutor{Requests: requests, Output: output, Grok: grokExecutor, Credentials: credentials, FastPolicy: fast, Continuation: &session.CompatResponses{TTL: choices.OpenAIHTTPResponseStickyTTL}, PromptCache: session.NewAnthropicPromptCache(time.Now), CodexUsage: aux.CodexUsage, ResponseTTL: choices.OpenAIHTTPResponseStickyTTL, Compact: &CompactExecutor{}}
	lineage := &OpenAIEncryptedLineage{Store: state, TTL: choices.SessionStickyTTL}
	imagePolicy := &gatewayadapter.ResponseImagePolicy{}
	responses := &OpenAIResponsesExecutor{Requests: requests, Output: output, Text: text, Grok: grokExecutor, Lineage: lineage, ImageBridge: imagePolicy}
	return &wsExecutionFixture{OpenAIResponsesExecutor: responses, Responses: responses, Text: text, choices: choices, options: v.options}
}

// newExecutionReadersFixture 测试读取器共享同一设置存储。
func newExecutionReadersFixture(repo settings.Repository, _ *wsFixtureOptions) *gatewayadapter.RuntimeReaders {
	if repo != nil {
		repo = settings.New(repo)
	}
	return gatewaytestkit.RuntimeReaders(repo)
}

func newUpstreamHealthForTest(store gatewayadapter.ExecutionProviderStore, _ *wsFixtureOptions, cache provider.TempUnschedCache, options provider.HealthOptions, readers *gatewayadapter.RuntimeReaders) *provideradapter.UpstreamHealth {
	return gatewaytestkit.NewHealthObserver(gatewaytestkit.HealthInput{Store: store, Cache: cache, Options: options, Readers: readers})
}

// setWSFixtureHealth 替换测试使用的健康观察接口，共用连接池和会话。
func setWSFixtureHealth(s *wsExecutionFixture, observer *provideradapter.UpstreamHealth) {
	s.Output.Health.Health = observer
	s.Output.GrokHealth.Health = observer
	s.Output.Observer = observer
}

// wsFixtureModelBlocked 按提供商健康接口的字段读取模型阻断状态。
func wsFixtureModelBlocked(s *wsExecutionFixture, value *gatewayadapter.ExecutionProvider, model string) bool {
	if value == nil {
		return false
	}
	key := provider.NormalizeTransientModel(gatewayadapter.ExecutionModelPolicy(value).CanonicalSchedulingModel(model))
	return s.Output.Health.ModelTransient.IsBlocked(value.Record.ID, key, time.Now())
}

func wsFixtureProviderBlocked(s *wsExecutionFixture, value *gatewayadapter.ExecutionProvider) bool {
	if value == nil || (value.Record.Platform != capability.PlatformOpenAI && value.Record.Platform != capability.PlatformGrok) {
		return false
	}
	return s.Output.Health.Runtime.Blocked(value.Record.ID, func() string { return provider.RefreshCredentialIdentity(gatewayadapter.ExecutionRecord(value)) })
}

func wsFixtureBlockProvider(s *wsExecutionFixture, value *gatewayadapter.ExecutionProvider, until time.Time, reason string) {
	if value == nil || (value.Record.Platform != capability.PlatformOpenAI && value.Record.Platform != capability.PlatformGrok) {
		return
	}
	s.Output.Health.Runtime.Block(value.Record.ID, until, reason)
}

func wsFixtureRetry429(s *wsExecutionFixture, value *gatewayadapter.ExecutionProvider, headers http.Header, body []byte) bool {
	return provideradapter.CanRetryOpenAI429(s.Output.Health.Runtime, gatewayadapter.ExecutionRecord(value), headers, body)
}

func wsFixtureRequestBlocked(s *wsExecutionFixture, value *gatewayadapter.ExecutionProvider, model string) bool {
	return wsFixtureProviderBlocked(s, value) || wsFixtureModelBlocked(s, value, model)
}
