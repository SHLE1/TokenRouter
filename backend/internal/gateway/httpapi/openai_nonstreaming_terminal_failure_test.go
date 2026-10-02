package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/upstream"

	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"

	providercore "github.com/TokenFlux/TokenRouter/internal/provider"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
)

// 这些测试检查 stream=false 请求收到 SSE 时的终止错误处理（#5281）。
// 兼容上游可能通过 HTTP 200 SSE 终止事件返回容量或限流错误。handleSSEToJSON 与 handlePassthroughSSEToJSON 使用流式读取器的错误分类决定是否换号。

func newNonStreamingFailoverContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	return c, rec
}

func newNonStreamingFailoverService() *OpenAIResponseOutput {
	return &OpenAIResponseOutput{Options: OpenAIResponseOptions{Configured: true, ReadLimit: 64 << 20}, Health: &provideradapter.OpenAIResponseHealth{Runtime: providercore.NewRuntimeBlockState(time.Now), ModelTransient: providercore.NewModelTransientState(0)}}
}

func newNonStreamingFailoverProvider() *gatewayprovider.ExecutionProvider {
	return &gatewayprovider.ExecutionProvider{
		Record: providercore.Record{
			LoadLocation: time.LoadLocation, ID: 1,
			Platform: capability.PlatformOpenAI,
			Type:     capability.ProviderTypeAPIKey,
			Name:     "pool-provider",
			Credentials: map[string]any{
				"pool_mode": true,
			},
		},
	}
}

func newNonStreamingSSEResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"text/event-stream"},
			"X-Request-Id": []string{"rid-nonstreaming-failed"},
		},
	}
}

func sseTerminalBody(eventType, data string) []byte {
	return []byte(strings.Join([]string{
		"event: " + eventType,
		"data: " + data,
		"",
		"data: [DONE]",
	}, "\n"))
}

// TestNonStreamingSSEToJSON_CapacityFailedEventFailsOver 验证非流式 SSE 的容量错误触发故障转移。
func TestNonStreamingSSEToJSON_CapacityFailedEventFailsOver(t *testing.T) {
	c, rec := newNonStreamingFailoverContext(t)
	svc := newNonStreamingFailoverService()
	body := sseTerminalBody("response.failed",
		`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`)

	result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), body, "model", "model")

	require.Nil(t, result)
	var failoverErr *forwardcore.UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	// invalid_request_error 按 400 分类，换号判断与流式路径一致。
	require.Equal(t, http.StatusBadRequest, failoverErr.StatusCode)
	// 容量降载按请求处理，先在同一提供商上有限次重试，与流式路径共用策略。
	require.True(t, failoverErr.RetryableOnSameProvider)
	require.Contains(t, string(failoverErr.ResponseBody), "Selected model is at capacity")
	// 换号的前提：一个字节都没写出去。
	require.False(t, c.Writer.Written())
	require.Empty(t, rec.Body.String())
}

// TestNonStreamingSSEToJSON_UnclassifiedFailedEventFailsOver 验证行为翻转点：未被分类为不可重试的泛化 response.failed，此前回 502，现在换号。
// 该结果与流式路径一致。
func TestNonStreamingSSEToJSON_UnclassifiedFailedEventFailsOver(t *testing.T) {
	c, rec := newNonStreamingFailoverContext(t)
	svc := newNonStreamingFailoverService()
	payload := []byte(`{"type":"response.failed","error":{"message":"upstream rejected request"}}`)
	body := sseTerminalBody("response.failed", string(payload))

	// 前提：流式分类器对同一帧的裁决就是「换号」。翻转不是新政策，是补齐。
	require.True(t, openai.OpenAIStreamFailedEventShouldFailover(payload, "upstream rejected request"))

	result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), body, "model", "model")

	require.Nil(t, result)
	var failoverErr *forwardcore.UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.False(t, c.Writer.Written())
	require.Empty(t, rec.Body.String())
}

// TestNonStreamingSSEToJSON_NonRetryableFailedEventStillWritesProtocolError 验证不可重试的 response.failed 返回 502 协议错误。
func TestNonStreamingSSEToJSON_NonRetryableFailedEventStillWritesProtocolError(t *testing.T) {
	cases := []struct {
		name    string
		data    string
		wantMsg string
	}{
		{
			name:    "invalid_request",
			data:    `{"type":"response.failed","error":{"type":"invalid_request_error","code":"invalid_request","message":"unknown parameter foo"}}`,
			wantMsg: "unknown parameter foo",
		},
		{
			name:    "context_window",
			data:    `{"type":"response.failed","response":{"id":"resp_failed","status":"failed","output":[],"error":{"code":"upstream_error","message":"input exceeds the context window"}}}`,
			wantMsg: "input exceeds the context window",
		},
		{
			name:    "content_policy",
			data:    `{"type":"response.failed","error":{"type":"content_policy_violation","message":"blocked by our content policy"}}`,
			wantMsg: "blocked by our content policy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newNonStreamingFailoverContext(t)
			svc := newNonStreamingFailoverService()

			result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), sseTerminalBody("response.failed", tc.data), "model", "model")

			require.Nil(t, result)
			require.Error(t, err)
			var failoverErr *forwardcore.UpstreamFailoverError
			require.False(t, errors.As(err, &failoverErr), "不可重试的上游错误不得换号")
			require.Equal(t, http.StatusBadGateway, rec.Code)
			require.Contains(t, rec.Body.String(), tc.wantMsg)
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		})
	}
}

// TestNonStreamingSSEToJSON_BareErrorEventUsesConservativeClassifier 验证裸 error 帧只有被识别为瞬时错误时才换号。
// error 与 response.failed 使用各自的分类器。
func TestNonStreamingSSEToJSON_BareErrorEventUsesConservativeClassifier(t *testing.T) {
	t.Run("non_transient_stays_protocol_error", func(t *testing.T) {
		c, rec := newNonStreamingFailoverContext(t)
		svc := newNonStreamingFailoverService()
		data := `{"type":"error","error":{"message":"upstream rejected request"}}`

		// 同一文案在 failed 分类中触发换号，在 error 分类中结束请求。
		require.True(t, openai.OpenAIStreamFailedEventShouldFailover([]byte(data), "upstream rejected request"))
		require.False(t, openai.OpenAIStreamErrorEventShouldFailover([]byte(data), "upstream rejected request"))

		result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), sseTerminalBody("error", data), "model", "model")

		require.Nil(t, result)
		require.Error(t, err)
		var failoverErr *forwardcore.UpstreamFailoverError
		require.False(t, errors.As(err, &failoverErr))
		require.Equal(t, http.StatusBadGateway, rec.Code)
	})

	t.Run("transient_fails_over", func(t *testing.T) {
		c, _ := newNonStreamingFailoverContext(t)
		svc := newNonStreamingFailoverService()
		data := `{"type":"error","error":{"message":"Temporary upstream failure, please retry"}}`

		result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), sseTerminalBody("error", data), "model", "model")

		require.Nil(t, result)
		var failoverErr *forwardcore.UpstreamFailoverError
		require.ErrorAs(t, err, &failoverErr)
		require.False(t, c.Writer.Written())
	})
}

// TestNonStreamingPassthroughSSEToJSON_CapacityFailedEventFailsOver 验证透传路径收到容量失败事件时换号。
func TestNonStreamingPassthroughSSEToJSON_CapacityFailedEventFailsOver(t *testing.T) {
	c, rec := newNonStreamingFailoverContext(t)
	svc := newNonStreamingFailoverService()
	body := sseTerminalBody("response.failed",
		`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`)

	result, err := openai.ReadPassthroughSSEAsJSON(newNonStreamingSSEResponse(), ResponseSink{Writer: c.Writer}, svc.PassthroughOptions(c.Request.Context(), c, newNonStreamingFailoverProvider()), body, "model", "model")

	require.Nil(t, result)
	var failoverErr *forwardcore.UpstreamFailoverError
	require.ErrorAs(t, err, &failoverErr)
	require.Contains(t, string(failoverErr.ResponseBody), "Selected model is at capacity")
	require.False(t, c.Writer.Written())
	require.Empty(t, rec.Body.String())
}

// TestNonStreamingSSEToJSON_MatchesStreamingClassifierVerdict 逐项对照非流式与流式错误分类结果。
func TestNonStreamingSSEToJSON_MatchesStreamingClassifierVerdict(t *testing.T) {
	payloads := []string{
		`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`,
		`{"type":"response.failed","error":{"message":"upstream rejected request"}}`,
		`{"type":"response.failed","error":{"type":"invalid_request_error","code":"invalid_request","message":"unknown parameter foo"}}`,
		`{"type":"response.failed","response":{"id":"r","status":"failed","output":[],"error":{"code":"upstream_error","message":"input exceeds the context window"}}}`,
		`{"type":"response.failed","error":{"type":"content_policy_violation","message":"blocked by our content policy"}}`,
	}

	for _, data := range payloads {
		t.Run(data[:min(len(data), 60)], func(t *testing.T) {
			payload := []byte(data)
			want := openai.OpenAIStreamFailedEventShouldFailover(payload, openai.ExtractOpenAISSEErrorMessage(payload))

			c, _ := newNonStreamingFailoverContext(t)
			svc := newNonStreamingFailoverService()
			_, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), sseTerminalBody("response.failed", data), "model", "model")

			var failoverErr *forwardcore.UpstreamFailoverError
			require.Equal(t, want, errors.As(err, &failoverErr),
				"非流式裁决与流式分类器不一致：%s", data)
		})
	}
}

// TestNonStreamingSSEToJSON_CommittedResponseKeepsProtocolError 验证已提交响应后返回协议错误。
// handler 的 openAIForwardMayFailover 按扣除心跳后的写出量决定是否换号（#3887）。
func TestNonStreamingSSEToJSON_CommittedResponseKeepsProtocolError(t *testing.T) {
	c, rec := newNonStreamingFailoverContext(t)
	svc := newNonStreamingFailoverService()
	MarkResponseCommitted(c)
	body := sseTerminalBody("response.failed",
		`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model.","type":"invalid_request_error"}}`)

	result, err := openai.ReadSSEAsJSON(context.Background(), newNonStreamingSSEResponse(), upstream.NewOutputContext(ResponseSink{Writer: c.Writer}), svc.NonStreamOptions(context.Background(), c, newNonStreamingFailoverProvider()), body, "model", "model")

	require.Nil(t, result)
	require.Error(t, err)
	var failoverErr *forwardcore.UpstreamFailoverError
	require.False(t, errors.As(err, &failoverErr))
	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestNonStreamingTerminalFailureFailover_NilProviderProposesNothing(t *testing.T) {
	c, _ := newNonStreamingFailoverContext(t)
	svc := newNonStreamingFailoverService()
	payload := []byte(`{"type":"response.failed","error":{"message":"Selected model is at capacity. Please try a different model."}}`)

	require.Nil(t, svc.nonStreamingTerminalFailure(
		c, newNonStreamingSSEResponse(), nil, false, "response.failed", payload,
		"Selected model is at capacity. Please try a different model."))
}
