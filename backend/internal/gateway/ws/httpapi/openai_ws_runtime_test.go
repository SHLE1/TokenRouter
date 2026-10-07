package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/gateway/session/testkit"
	"github.com/TokenFlux/TokenRouter/internal/gateway/ws"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient/tlsfingerprint"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// runtimeTestDialer 记录拨号预算，并为后台预热创建独立连接。
type runtimeTestDialer struct {
	mu      sync.Mutex
	conns   []openai.WSClientConn
	budgets []time.Duration
}

func (d *runtimeTestDialer) Dial(ctx context.Context, _ string, _ http.Header, _ string, _ *tlsfingerprint.Profile) (openai.WSClientConn, int, http.Header, error) {
	deadline, _ := ctx.Deadline()
	d.mu.Lock()
	n := len(d.budgets)
	d.budgets = append(d.budgets, time.Until(deadline))
	var conn openai.WSClientConn
	if n < len(d.conns) {
		conn = d.conns[n]
	}
	d.mu.Unlock()
	if conn != nil {
		return conn, 0, nil, nil
	}
	// 延迟预热拨号，使其与后续客户端轮次交错执行。
	time.Sleep(time.Millisecond)
	return &openAIWSFakeConn{}, 0, nil, nil
}

// TestWSRuntimeAcrossTurns 覆盖后台预热与逐轮快照并发，以及热更新后的重连预算。
func TestWSRuntimeAcrossTurns(t *testing.T) {
	previousPingIdle := openAIWSIngressPreflightPingIdle
	openAIWSIngressPreflightPingIdle = 0
	defer func() { openAIWSIngressPreflightPingIdle = previousPingIdle }()
	for _, reconnect := range []bool{false, true} {
		name := "prewarm"
		if reconnect {
			name = "reconnect_timeout"
		}
		t.Run(name, func(t *testing.T) {
			first := []byte(`{"type":"response.completed","response":{"id":"resp_first","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`)
			second := []byte(`{"type":"response.completed","response":{"id":"resp_second","model":"gpt-5.1","usage":{"input_tokens":1,"output_tokens":1}}}`)
			dialer := &runtimeTestDialer{conns: []openai.WSClientConn{&openAIWSCaptureConn{events: [][]byte{first, second}}}}
			options := &wsFixtureOptions{}
			options.Pool.MaxConnsPerProvider = 4
			options.Pool.MinIdlePerProvider = 4
			options.Pool.MaxIdlePerProvider = 4
			options.WS.DialTimeoutSeconds = 1
			if reconnect {
				options.Pool.MaxConnsPerProvider = 1
				options.Pool.MinIdlePerProvider = 0
				options.Pool.MaxIdlePerProvider = 1
				dialer.conns = []openai.WSClientConn{&openAIWSPreflightFailConn{events: [][]byte{first}}, &openAIWSCaptureConn{events: [][]byte{second}}}
			}
			pool := newOpenAIWSConnPool(options)
			pool.SetClientDialerForTest(dialer)
			defer pool.Close()
			svc := newWSFixture(wsFixtureInputs{options: options, pool: pool, transport: &auxiliaryHTTPRecorder{}, cache: &testkit.StickyCache{}})
			defaults := ws.DefaultParameters()
			defaults.DialTimeoutSeconds = 1
			svc.Runtime = ws.NewRuntime(defaults, nil, nil)
			svc.Runtime.SetApply(func(value ws.Parameters) error {
				updated := *wsFixturePoolOptions(options)
				updated.DialTimeoutSeconds = value.DialTimeoutSeconds
				return pool.UpdateOptions(updated)
			})
			provider := &gatewayprovider.ExecutionProvider{Record: provider.Record{
				LoadLocation: time.LoadLocation, ID: 119, Platform: capability.PlatformOpenAI,
				Type: capability.ProviderTypeAPIKey, Concurrency: 1,
				Credentials: map[string]any{"api_key": "test"},
			}}
			finished := make(chan error, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					finished <- err
					return
				}
				defer func() { _ = conn.CloseNow() }()
				ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
				defer cancel()
				_, payload, err := conn.Read(ctx)
				if err != nil {
					finished <- err
					return
				}
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = r
				finished <- svc.ProxyResponsesWebSocketFromClient(ctx, c, conn, provider, "test", payload, nil)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			client, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
			require.NoError(t, err)
			defer func() { _ = client.CloseNow() }()
			turn := func(payload, responseID string) {
				require.NoError(t, client.Write(ctx, websocket.MessageText, []byte(payload)))
				_, event, err := client.Read(ctx)
				require.NoError(t, err)
				require.Equal(t, responseID, gjson.GetBytes(event, "response.id").String())
			}
			turn(`{"type":"response.create","model":"gpt-5.1"}`, "resp_first")
			require.NoError(t, svc.Runtime.Publish(`{"dial_timeout_seconds":10}`))
			turn(`{"type":"response.create","model":"gpt-5.1","previous_response_id":"resp_first"}`, "resp_second")
			require.NoError(t, client.Close(websocket.StatusNormalClosure, ""))
			require.NoError(t, <-finished)
			if reconnect {
				dialer.mu.Lock()
				budgets := append([]time.Duration(nil), dialer.budgets...)
				dialer.mu.Unlock()
				require.Len(t, budgets, 2)
				require.InDelta(t, 3, budgets[0].Seconds(), 0.5)
				require.InDelta(t, 12, budgets[1].Seconds(), 0.5)
			}
		})
	}
}
