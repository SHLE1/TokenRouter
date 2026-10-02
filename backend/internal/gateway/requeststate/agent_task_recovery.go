package requeststate

import "context"

type agentTaskRecoveryKey struct{}

// WithAgentTaskRecovery 标记当前尝试序列已使用恢复机会。
func WithAgentTaskRecovery(ctx context.Context) context.Context {
	return context.WithValue(ctx, agentTaskRecoveryKey{}, true)
}

func AgentTaskRecoveryTried(ctx context.Context) bool {
	value, _ := ctx.Value(agentTaskRecoveryKey{}).(bool)
	return value
}
