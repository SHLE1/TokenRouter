package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/gateway/errorpolicy"
	"github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	"github.com/TokenFlux/TokenRouter/internal/moderation"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type localizedErrorRule struct {
	rule *errorpolicy.ErrorPassthroughRule
}

func (r localizedErrorRule) MatchRule(string, int, []byte) *errorpolicy.ErrorPassthroughRule {
	return r.rule
}

// TestLocalizedHTTPAndSSEErrors 检查自定义译文与上游原文的输出路径。
func TestLocalizedHTTPAndSSEErrors(t *testing.T) {
	en := "en"
	message := "Original"
	rule := &errorpolicy.ErrorPassthroughRule{PassthroughCode: true, CustomMessage: &message, MessageLocalization: errorpolicy.MessageLocalization{SourceLocale: &en, Source: message, Revision: 1, SourceRevision: 1, Translations: map[string]locale.Translation[string]{"zh-Hans": {Value: "译文", SourceRevision: 1}}}}
	for _, stream := range []bool{false, true} {
		for _, language := range []string{"en", "zh-Hans"} {
			for _, raw := range []bool{false, true} {
				t.Run(language, func(t *testing.T) {
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil).WithContext(locale.WithLanguage(context.Background(), language))
					rule.PassthroughBody = raw
					WriteAnthropicFailover(c, &forward.UpstreamFailoverError{StatusCode: 429, ResponseBody: []byte(`{"error":{"message":"Upstream evidence"}}`)}, "anthropic", stream, localizedErrorRule{rule}, func([]byte) bool { return false }, "")
					want := message
					if language == "zh-Hans" {
						want = "译文"
					}
					if raw {
						want = "Upstream evidence"
					}
					require.Contains(t, recorder.Body.String(), want)
					require.Contains(t, recorder.Body.String(), `"type":"upstream_error"`)
					if stream {
						require.Equal(t, http.StatusOK, recorder.Code)
						require.True(t, strings.HasPrefix(recorder.Body.String(), "data: "))
					} else {
						require.Equal(t, http.StatusTooManyRequests, recorder.Code)
					}
				})
			}
		}
	}
	status, code, translated, _ := BillingErrorDetails(billing.ErrAPIKeyRateLimit5hExceeded, "en")
	require.Equal(t, 429, status)
	require.Equal(t, "rate_limit_exceeded", code)
	require.Equal(t, "The API key five-hour limit has been reached.", translated)
}

// TestLocalizedWebSocketModeration 检查 WebSocket 错误事件的语言与固定错误码。
func TestLocalizedWebSocketModeration(t *testing.T) {
	for _, language := range []string{"en", "zh-Hans"} {
		t.Run(language, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					return
				}
				defer func() { _ = conn.CloseNow() }()
				WriteResponsesWSModeration(locale.WithLanguage(r.Context(), language), conn, &moderation.Decision{})
			}))
			defer server.Close()
			conn, _, err := websocket.Dial(context.Background(), strings.Replace(server.URL, "http:", "ws:", 1), nil)
			require.NoError(t, err)
			defer func() { _ = conn.CloseNow() }()
			kind, body, err := conn.Read(context.Background())
			require.NoError(t, err)
			require.Equal(t, websocket.MessageText, kind)
			var event struct {
				Type  string `json:"type"`
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.Unmarshal(body, &event))
			require.Equal(t, "error", event.Type)
			require.Equal(t, "content_policy_violation", event.Error.Code)
			if language == "en" {
				require.Equal(t, "content moderation blocked this request", event.Error.Message)
			} else {
				require.Equal(t, "此请求已被内容审核阻止。", event.Error.Message)
			}
		})
	}
}
