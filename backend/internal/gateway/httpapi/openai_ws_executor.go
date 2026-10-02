package httpapi

import (
	"context"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/gateway/promptpolicy"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session"
)

// OpenAIWSOptions 保存 WS 执行的静态参数，nil 表示使用默认配置。
type OpenAIWSOptions struct {
	Enabled, OAuthEnabled, APIKeyEnabled, ForceHTTP                      bool
	ResponsesWebsockets, ResponsesWebsocketsV2, ModeRouterV2Enabled      bool
	IngressModeDefault                                                   string
	ClientFirstMessageTimeoutSeconds, IngressInterTurnIdleTimeoutSeconds int
	ClientReadLimitBytes, HTTPBridgeThresholdBytes                       int64
	HTTPBridgeEnabled                                                    bool
	AllowStoreRecovery, IngressPreviousResponseRecoveryEnabled           bool
	StoreDisabledConnMode                                                string
	StoreDisabledForceNewConn, PrewarmGenerateEnabled                    bool
	DialTimeoutSeconds, ReadTimeoutSeconds, WriteTimeoutSeconds          int
	EventFlushBatchSize, EventFlushIntervalMS, PrewarmCooldownMS         int
	RetryBackoffInitialMS, RetryBackoffMaxMS, RetryTotalBudgetMS         int
	RetryJitterRatio, PayloadLogSampleRate                               float64
	StickyResponseIDTTLSeconds                                           int
}

// OpenAIWSSelection 只提供已经选中提供商的传输选择与会话预算。
type OpenAIWSSelection interface {
	ResolveTransport(*gatewayadapter.ExecutionProvider) egress.OpenAIWSProtocolDecision
	SessionStickyTTL() time.Duration
}

// OpenAIWSDependencies 保存 WS 执行使用的接口。
type OpenAIWSDependencies struct {
	Options     *OpenAIWSOptions
	Connections *OpenAIWSConnections
	Requests    *OpenAIRequests
	Output      *OpenAIResponseOutput
	Grok        *GrokExecutor
	FastPolicy  *gatewayadapter.ExecutionFastPolicy
	Prompts     *promptpolicy.Service
	Selection   OpenAIWSSelection
	State       session.OpenAIWSStateStore
	Lineage     *OpenAIEncryptedLineage
	ImageBridge *gatewayadapter.ResponseImagePolicy
	Cache       session.GatewayCache
}

// OpenAIWebSocketExecutor 适配帧、凭据、健康与HTTP桥接，逐轮循环由gateway/ws唯一拥有。
type OpenAIWebSocketExecutor struct {
	OpenAIWSDependencies
	openaiWSSessionPreemptions openAIWSSessionPreemptRegistry
	openaiWSRetryMetrics       openAIWSRetryMetrics
}

// NewOpenAIWebSocketExecutor 构造不启动连接或会话任务。
func NewOpenAIWebSocketExecutor(deps OpenAIWSDependencies) *OpenAIWebSocketExecutor {
	out := &OpenAIWebSocketExecutor{OpenAIWSDependencies: deps}
	out.logOpenAIWSModeBootstrap()
	return out
}

// EnsureSessionIsolation 使用当前请求的认证 Key 检查会话隔离。
func (s *OpenAIWebSocketExecutor) EnsureSessionIsolation(ctx context.Context, key *apikey.APIKey, userID int64, source, hash string) error {
	if key == nil {
		return nil
	}
	var group int64
	if key.GroupID != nil {
		group = *key.GroupID
	}
	return session.EnsureIsolation(ctx, s.Cache, session.IsolationInput{UserID: userID, GroupID: group, Source: source, Hash: hash, TTL: time.Hour, Enabled: key.Group != nil && key.Group.SessionIsolationEnabled})
}
