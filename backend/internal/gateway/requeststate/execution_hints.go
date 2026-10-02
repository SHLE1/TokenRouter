package requeststate

import "context"

// Hint 通过设置标记区分未提供与零值，保存传入数据的值副本。
type Hint[T bool | int | int64] struct {
	Value T
	Set   bool
}

// ExecutionHints 是本次请求的执行参数快照；派生 attempt 不修改父请求。
type ExecutionHints struct {
	// HealthModel 保存本次尝试规范化后的型号，健康观测直接使用该值。
	HealthModel            string
	SessionIsolationSource string
	SessionIsolationHash   string
	// 选择快照保存本次阈值、分组要求和单次降级标记。
	QuotaAutoPauseThreshold5h   float64
	QuotaAutoPauseThreshold7d   float64
	GroupPrivacyRequirement     Hint[bool]
	GroupPrivacyGroupID         int64
	ProxyQuarantineBypass       bool
	ClaudeCode                  bool
	ClaudeCodeVersion           string
	OpenAIImageGenerationIntent bool
	OpenAIImagesEndpoint        bool
	IsMaxTokensOneHaikuRequest  Hint[bool]
	ThinkingEnabled             Hint[bool]
	PrefetchedStickyProviderID  Hint[int64]
	PrefetchedStickyGroupID     Hint[int64]
	SingleProviderRetry         Hint[bool]
	ProviderSwitchCount         Hint[int]
}

// IsClaudeCodeClient 返回入口保存的 Claude Code 客户端判断结果。
func IsClaudeCodeClient(ctx context.Context) bool {
	return ExecutionHintsFromContext(ctx).ClaudeCode
}

func SetClaudeCodeClient(ctx context.Context, value bool) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.ClaudeCode = value })
}

func SetClaudeCodeVersion(ctx context.Context, value string) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.ClaudeCodeVersion = value })
}

func GetClaudeCodeVersion(ctx context.Context) string {
	return ExecutionHintsFromContext(ctx).ClaudeCodeVersion
}

func WithOpenAIImageGenerationIntent(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return updateHints(ctx, func(h *ExecutionHints) { h.OpenAIImageGenerationIntent = true })
}

func OpenAIImageGenerationIntentFromContext(ctx context.Context) bool {
	return ExecutionHintsFromContext(ctx).OpenAIImageGenerationIntent
}

func WithOpenAIImagesEndpoint(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return updateHints(ctx, func(h *ExecutionHints) { h.OpenAIImagesEndpoint = true })
}

func OpenAIImagesEndpointFromContext(ctx context.Context) bool {
	return ExecutionHintsFromContext(ctx).OpenAIImagesEndpoint
}

type executionHintsKey struct{}

// ExecutionHintsFromContext 读取 context 中的执行提示，平台执行入口通过参数接收该值。
func ExecutionHintsFromContext(ctx context.Context) ExecutionHints {
	if ctx == nil {
		return ExecutionHints{}
	}
	hints, _ := ctx.Value(executionHintsKey{}).(ExecutionHints)
	return hints
}

func WithExecutionHints(ctx context.Context, hints ExecutionHints) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, executionHintsKey{}, hints)
}

func updateHints(ctx context.Context, update func(*ExecutionHints)) context.Context {
	hints := ExecutionHintsFromContext(ctx)
	update(&hints)
	return WithExecutionHints(ctx, hints)
}

func WithIsMaxTokensOneHaikuRequest(ctx context.Context, value bool) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.IsMaxTokensOneHaikuRequest = Hint[bool]{value, true} })
}

func WithThinkingEnabled(ctx context.Context, value bool) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.ThinkingEnabled = Hint[bool]{value, true} })
}

func WithPrefetchedStickySession(ctx context.Context, providerID, groupID int64) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) {
		h.PrefetchedStickyProviderID = Hint[int64]{providerID, true}
		h.PrefetchedStickyGroupID = Hint[int64]{groupID, true}
	})
}

func WithSingleProviderRetry(ctx context.Context, value bool) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.SingleProviderRetry = Hint[bool]{value, true} })
}

func WithProviderSwitchCount(ctx context.Context, value int) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.ProviderSwitchCount = Hint[int]{value, true} })
}

func IsMaxTokensOneHaikuRequestFromContext(ctx context.Context) (bool, bool) {
	h := ExecutionHintsFromContext(ctx).IsMaxTokensOneHaikuRequest
	return h.Value, h.Set
}

func ThinkingEnabledFromContext(ctx context.Context) (bool, bool) {
	h := ExecutionHintsFromContext(ctx).ThinkingEnabled
	return h.Value, h.Set
}

func PrefetchedStickyProviderIDFromContext(ctx context.Context) (int64, bool) {
	h := ExecutionHintsFromContext(ctx).PrefetchedStickyProviderID
	return h.Value, h.Set
}

func PrefetchedStickyGroupIDFromContext(ctx context.Context) (int64, bool) {
	h := ExecutionHintsFromContext(ctx).PrefetchedStickyGroupID
	return h.Value, h.Set
}

func SingleProviderRetryFromContext(ctx context.Context) (bool, bool) {
	h := ExecutionHintsFromContext(ctx).SingleProviderRetry
	return h.Value, h.Set
}

func ProviderSwitchCountFromContext(ctx context.Context) (int, bool) {
	h := ExecutionHintsFromContext(ctx).ProviderSwitchCount
	return h.Value, h.Set
}

// WithSessionIsolation 保存入口解析出的指定会话身份，供分组回退时再次校验。
func WithSessionIsolation(ctx context.Context, source, hash string) context.Context {
	return updateHints(ctx, func(h *ExecutionHints) { h.SessionIsolationSource, h.SessionIsolationHash = source, hash })
}
