package httpapi

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	"github.com/TokenFlux/TokenRouter/internal/protocol/openai"
	"github.com/google/uuid"
)

// completeHTTPBridgeWarmup 在本地完成预热，入站会话继续保存输入供下一轮重放。
func completeHTTPBridgeWarmup(model string, write func([]byte) error) (*forward.OpenAIResult, error) {
	started := time.Now()
	responseID := "resp_ws_warmup_" + uuid.NewString()
	events, err := openai.WSWarmupEvents(responseID, model, started.Unix())
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if err := write(event); err != nil {
			return nil, err
		}
	}
	return &forward.OpenAIResult{
		LocalWarmup: true, RequestID: responseID, ResponseID: responseID, Model: model,
		Stream: true, OpenAIWSMode: true, UpstreamTerminalEvent: "response.completed", Duration: time.Since(started),
	}, nil
}
