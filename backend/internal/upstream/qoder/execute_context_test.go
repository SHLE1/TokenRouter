package qoder

import (
	"context"
	"net/http"
	"testing"
	"testing/synctest"

	"github.com/TokenFlux/TokenRouter/internal/pkg/requestcontext"
	"github.com/TokenFlux/TokenRouter/internal/protocol"
	"github.com/TokenFlux/TokenRouter/internal/upstream"
	"github.com/stretchr/testify/require"
)

// contextTestClient 在进入推理后检查传给供应商的取消信号。
type contextTestClient struct{ check func(context.Context) error }

func (c contextTestClient) StreamRequestContext(ctx context.Context, _ *SessionContext, _ string, _ []byte, _ map[string]string) (*http.Response, error) {
	return nil, c.check(ctx)
}

// TestStreamExecutionInternalAbort 验证三种协议都保留内部终止并隔离客户端断连。
func TestStreamExecutionInternalAbort(t *testing.T) {
	for _, wire := range []protocol.ProtocolID{protocol.ProtocolOpenAIChatCompletions, protocol.ProtocolOpenAIResponses, protocol.ProtocolAnthropicMessages} {
		t.Run(string(wire), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				client, cancelClient := context.WithCancel(context.Background())
				request, abort := requestcontext.WithAbort(client)
				defer abort()
				called := false
				target := &Target{
					ProviderID: 1,
					Session:    func(context.Context) (*SessionContext, error) { return &SessionContext{}, nil },
					Client: func() (StreamClient, error) {
						return contextTestClient{check: func(ctx context.Context) error {
							called = true
							cancelClient()
							synctest.Wait()
							require.NoError(t, ctx.Err())
							abort()
							synctest.Wait()
							require.ErrorIs(t, ctx.Err(), context.Canceled)
							return ctx.Err()
						}}, nil
					},
				}
				_, err := NewExecutor(ExecuteOptions{}).Execute(request, upstream.AttemptInput{
					Target: target, Protocol: wire, Stream: true,
					Body: []byte(`{"model":"auto","input":"hi","messages":[{"role":"user","content":"hi"}],"max_tokens":64}`),
				}, &modelTestSink{})
				require.True(t, called)
				require.ErrorIs(t, err, context.Canceled)
			})
		})
	}
}
