package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/identity/httpapi/authctx"

	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	httptestkit "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi/testkit"
	"github.com/TokenFlux/TokenRouter/internal/infra/telemetry/logging"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/billing"
	"github.com/TokenFlux/TokenRouter/internal/config"

	"github.com/TokenFlux/TokenRouter/internal/identity"

	"github.com/TokenFlux/TokenRouter/internal/routing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 非法 service_tier 在 /v1/responses 和 /v1/chat/completions 入口
// 返回 OpenAI 格式的 HTTP 400 错误，handler 在字段校验时结束请求。
//
// 合法值（fast/priority/flex/auto/default/scale/ultrafast）及省略、null 的处理
// 由纯校验测试 TestValidateOpenAIServiceTierField 覆盖。

func newServiceTierHandlerTest(t *testing.T) *gatewayHTTPEndpointsFixture {
	t.Helper()
	return newGatewayHTTPEndpoints(gatewayHTTPFixtureInput{
		Source:  &gatewayExecutionFixture{},
		Funding: newFundingAdmissionFixture(newBillingEligibilityFixture(&config.Config{}), &config.Config{}),
		Keys:    &apikey.APIKeyService{},
		Concurrency: gatewayhttp.NewConcurrencyHelper(scheduler.NewConcurrencyService(
			&httptestkit.ConcurrencySequence{UserSeq: []bool{true}}, scheduler.Diagnostics{
				Logf:  logging.LegacyPrintf,
				Event: logging.Event,
			},
		), gatewayhttp.SSEPingFormatNone, 0),
		Config: &config.Config{},
		Images: &scheduler.ImageConcurrencyLimiter{}, Availability: newExecutionAvailabilityForTest(nil, nil, nil), Choices: newEmptyCompatibleSelectionFixture(),
	})
}

func runOpenAIHandlerServiceTierTest(t *testing.T, path, body string, handler func(h *gatewayHTTPEndpointsFixture, c *gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	groupID := int64(6401)
	userID := int64(6402)
	c.Set(string(keyhttp.ContextKeyAPIKey), &apikey.APIKey{
		ID:      6403,
		GroupID: &groupID,
		Group: &routing.Group{
			ID: groupID,
		},
		User: &identity.User{ID: userID, Status: billing.StatusActive},
	})
	c.Set(string(authctx.ContextKeyUser), authctx.AuthSubject{UserID: userID, Concurrency: 1})

	handler(newServiceTierHandlerTest(t), c)
	return rec
}

func TestOpenAIGatewayHandlerResponses_InvalidServiceTierRejected400(t *testing.T) {
	for _, body := range []string{
		`{"model":"gpt-5.5","input":"hi","service_tier":"turbo"}`,
		`{"model":"gpt-5.5","input":"hi","service_tier":"SPEED"}`,
		`{"model":"gpt-5.5","input":"hi","service_tier":""}`,
		`{"model":"gpt-5.5","input":"hi","service_tier":123}`,
		`{"model":"gpt-5.5","input":"hi","service_tier":{}}`,
	} {
		rec := runOpenAIHandlerServiceTierTest(t, "/v1/responses", body, func(h *gatewayHTTPEndpointsFixture, c *gin.Context) {
			h.Responses(c)
		})
		require.Equal(t, http.StatusBadRequest, rec.Code, "body=%s", body)
		require.Contains(t, rec.Body.String(), "invalid_request_error", "body=%s", body)
		require.Contains(t, rec.Body.String(), "invalid service_tier", "body=%s", body)
	}
}

func TestOpenAIGatewayHandlerChatCompletions_InvalidServiceTierRejected400(t *testing.T) {
	for _, body := range []string{
		`{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}],"service_tier":"turbo"}`,
		`{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}],"service_tier":"ultra"}`,
		`{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}],"service_tier":""}`,
		`{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}],"service_tier":["priority"]}`,
	} {
		rec := runOpenAIHandlerServiceTierTest(t, "/v1/chat/completions", body, func(h *gatewayHTTPEndpointsFixture, c *gin.Context) {
			h.ChatCompletions(c)
		})
		require.Equal(t, http.StatusBadRequest, rec.Code, "body=%s", body)
		require.Contains(t, rec.Body.String(), "invalid_request_error", "body=%s", body)
		require.Contains(t, rec.Body.String(), "invalid service_tier", "body=%s", body)
	}
}
