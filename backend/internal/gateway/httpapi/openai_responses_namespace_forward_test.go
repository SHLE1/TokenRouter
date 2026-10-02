package httpapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/protocol/bridge"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// codexNamespaceRequestBody 模拟 Codex 多智能体请求及带残留 namespace 的普通消息项。
const codexNamespaceRequestBody = `{
	"model":"gpt-5.6-terra",
	"stream":false,
	"instructions":"test",
	"tools":[
		{"type":"namespace","name":"collaboration","description":"Tools for spawning and managing sub-agents.","tools":[
			{"type":"function","name":"spawn_agent","description":"Call as to=functions.collaboration.spawn_agent","parameters":{"type":"object"}},
			{"type":"function","name":"wait_agent","parameters":{"type":"object"}}
		]},
		{"type":"function","name":"exec","parameters":{"type":"object"}}
	],
	"input":[
		{"type":"function_call","namespace":"collaboration","name":"spawn_agent","call_id":"call_1","arguments":"{}"},
		{"type":"message","role":"user","namespace":"leftover","content":[{"type":"input_text","text":"hello"}]}
	]
}`

const namespaceForwardOKResponse = `{"id":"resp_ns","output":[],"usage":{"input_tokens":1,"output_tokens":1,"input_tokens_details":{"cached_tokens":0}}}`

// TestOpenAIGatewayService_OAuthPreservesCodexNamespaceTools 验证 OAuth Responses 原样保留 namespace 声明和历史工具调用字段。
func TestOpenAIGatewayService_OAuthPreservesCodexNamespaceTools(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &auxiliaryHTTPRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	c := newOpenAIRejectedFieldTestContext(body)

	result, err := newOpenAIRejectedFieldTestService(upstream).Forward(
		context.Background(), c, newOpenAIOAuthNamespaceTestProvider(), body,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.bodies, 1)
	forwarded := upstream.bodies[0]
	namespaceTool := gjson.GetBytes(forwarded, `tools.#(type=="namespace")`)
	require.True(t, namespaceTool.Exists())
	require.Equal(t, "collaboration", namespaceTool.Get("name").String())
	require.Equal(t, "spawn_agent", namespaceTool.Get("tools.0.name").String())
	require.NotContains(t, string(forwarded), "collaboration__spawn_agent")
	require.Equal(t, "collaboration", gjson.GetBytes(forwarded, "input.0.namespace").String())
	require.False(t, gjson.GetBytes(forwarded, "input.1.namespace").Exists())
	require.Empty(t, OpenAIResponsesNamespaceNames(c))
}

// TestOpenAIGatewayService_APIKeyPreservesDeclaredNamespaceToolCalls 验证API Key 自定义上游若接受 namespace 工具声明，也要求历史 function_call 原样携带
// namespace。声明仍为命名空间工具却清掉调用项字段，会触发 Missing namespace。
func TestOpenAIGatewayService_APIKeyPreservesDeclaredNamespaceToolCalls(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &auxiliaryHTTPRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	c := newOpenAIRejectedFieldTestContext(body)

	result, err := newOpenAIRejectedFieldTestService(upstream).Forward(
		context.Background(), c, newOpenAIRejectedFieldTestProvider(), body,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.bodies, 1)
	forwarded := upstream.bodies[0]

	require.True(t, gjson.GetBytes(forwarded, `tools.#(type=="namespace")`).Exists())
	require.Equal(t, "collaboration", gjson.GetBytes(forwarded, "input.0.namespace").String())
	require.False(t, gjson.GetBytes(forwarded, "input.1.namespace").Exists())
}

// TestOpenAIGatewayService_OAuthCompactKeepsFlattening 验证 Compact 请求摊平 namespace 并清理工具声明。
// Compact 执行历史摘要，input[].namespace 会触发 400 Unknown parameter（#4761）。
func TestOpenAIGatewayService_OAuthCompactKeepsFlattening(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &auxiliaryHTTPRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	c := newOpenAIRejectedFieldTestContext(body)
	c.Request.URL.Path = "/v1/responses/compact"

	result, err := newOpenAIRejectedFieldTestService(upstream).Forward(
		context.Background(), c, newOpenAIOAuthNamespaceTestProvider(), body,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	forwarded := upstream.bodies[0]
	require.False(t, gjson.GetBytes(forwarded, "input.0.namespace").Exists())
	require.False(t, gjson.GetBytes(forwarded, `tools.#(type=="namespace")`).Exists())
	require.Equal(t, "collaboration__spawn_agent", gjson.GetBytes(forwarded, "input.0.name").String())
}

// TestOpenAIGatewayService_OAuthFlattenFlagRestoresLegacyBehavior 验证提供商兼容开关打开后恢复 namespace 摊平旧行为。
func TestOpenAIGatewayService_OAuthFlattenFlagRestoresLegacyBehavior(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &auxiliaryHTTPRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	c := newOpenAIRejectedFieldTestContext(body)
	provider := newOpenAIOAuthNamespaceTestProvider()
	provider.Record.Extra = map[string]any{"openai_responses_flatten_namespaces": true}

	result, err := newOpenAIRejectedFieldTestService(upstream).Forward(context.Background(), c, provider, body)

	require.NoError(t, err)
	require.NotNil(t, result)
	forwarded := upstream.bodies[0]
	require.False(t, gjson.GetBytes(forwarded, `tools.#(type=="namespace")`).Exists())
	require.True(t, gjson.GetBytes(forwarded, `tools.#(name=="collaboration__spawn_agent")`).Exists())
	require.False(t, gjson.GetBytes(forwarded, "input.0.namespace").Exists())
	require.Equal(t, bridge.ResponsesNamespaceName{
		Namespace: "collaboration",
		Name:      "spawn_agent",
	}, OpenAIResponsesNamespaceNames(c)["collaboration__spawn_agent"])
}

// TestOpenAIGatewayService_ForwardClearsStaleNamespaceNames 验证 failover 复用 gin.Context 时，转发前清除上个提供商的映射。
func TestOpenAIGatewayService_ForwardClearsStaleNamespaceNames(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &auxiliaryHTTPRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	c := newOpenAIRejectedFieldTestContext(body)
	SetOpenAIResponsesNamespaceNames(c, map[string]bridge.ResponsesNamespaceName{
		"stale__tool": {Namespace: "stale", Name: "tool"},
	})

	_, err := newOpenAIRejectedFieldTestService(upstream).Forward(
		context.Background(), c, newOpenAIOAuthNamespaceTestProvider(), body,
	)

	require.NoError(t, err)
	require.Empty(t, OpenAIResponsesNamespaceNames(c))
}
