package httpapi

import (
	"context"
	"sync"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/provider"

	"github.com/TokenFlux/TokenRouter/internal/gateway/completion"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai/liveattestation"
)

// OpenAILiveOptions 保存构造时的会话预算及观察重试间隔。
type OpenAILiveOptions struct {
	MaxSessionDuration    time.Duration
	ObserverRetryInterval time.Duration
}

// LiveProviderSelection 提供创建会话时的提供商选择和每轮资格复核。
type LiveProviderSelection interface {
	SelectProviderWithSchedulerForCapability(context.Context, *int64, string, string, string, map[int64]struct{}, egress.OpenAIUpstreamTransport, provider.OpenAIEndpointCapability, bool, bool, ...string) (*gatewayadapter.SelectionResult, scheduler.PlatformDecision, error)
	ResolveOpenAIWSRoutingModelForProvider(context.Context, *int64, *gatewayadapter.ExecutionProvider, string, provider.OpenAIEndpointCapability) (string, error)
}

// OpenAILiveExecutor 管理 Live 依赖和观察者登记，共用会话、租约与传输。
type OpenAILiveExecutor struct {
	Options             OpenAILiveOptions
	Requests            *OpenAIRequests
	Selection           LiveProviderSelection
	Routes              *gatewayadapter.RoutePlanner
	Store               session.LiveCallStore
	Leases              scheduler.LiveConcurrencyCache
	Usage               *completion.Recorder
	Dialer              openai.WSClientDialer
	Attestation         liveattestation.Provider
	AttestationCipher   identity.SecretEncryptor
	Background          func(string, func()) bool
	liveObserverMu      sync.Mutex
	liveObserverStopped bool
	liveObserverCancels map[string]context.CancelFunc
	liveObserverWG      sync.WaitGroup
}
