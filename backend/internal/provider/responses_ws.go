package provider

import (
	"maps"
	"slices"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

const (
	ResponsesWSConnectionModeKey = "responses_ws_connection_mode"
	ResponsesWSPooled            = "pooled"
	ResponsesWSPerSession        = "per_session"
)

// ResponsesWSConnectionMode 返回已保存的连接方式，缺省时复用连接。
func (a *Record) ResponsesWSConnectionMode() string {
	if a != nil && a.Extra[ResponsesWSConnectionModeKey] == ResponsesWSPerSession {
		return ResponsesWSPerSession
	}
	return ResponsesWSPooled
}

// SupportsResponsesWS 根据上游协议集合判断是否可以直接建立 Responses WS。
func (a *Record) SupportsResponsesWS() bool {
	return a != nil && a.IsOpenAI() && slices.Contains(a.UpstreamProtocols(), capability.ProtocolResponsesWebSocket)
}

// NormalizeResponsesWS 在保存和导入时转换旧连接设置，分组决定客户端协议许可。
func NormalizeResponsesWS(a *Record) error {
	if a == nil || !a.IsOpenAI() {
		return nil
	}
	extra := maps.Clone(a.Extra)
	if extra == nil {
		extra = map[string]any{}
	}
	mode := ResponsesWSPooled
	bridge := false
	if raw, exists := extra[ResponsesWSConnectionModeKey]; exists {
		value, ok := raw.(string)
		if !ok || (value != ResponsesWSPooled && value != ResponsesWSPerSession) {
			return apperror.BadRequest("INVALID_RESPONSES_WS_CONNECTION_MODE", "responses_ws_connection_mode must be pooled or per_session")
		}
		mode = value
	} else {
		key := "openai_oauth_responses_websockets_v2_mode"
		if a.IsOpenAIApiKey() {
			key = "openai_apikey_responses_websockets_v2_mode"
		}
		legacy, _ := extra[key].(string)
		switch strings.ToLower(strings.TrimSpace(legacy)) {
		case "passthrough":
			mode = ResponsesWSPerSession
		case "http_bridge":
			bridge = true
		}
		if forced, _ := extra["openai_ws_force_http"].(bool); forced {
			bridge = true
		}
	}
	if bridge {
		protocols := slices.Clone(a.UpstreamProtocols())
		protocols = slices.DeleteFunc(protocols, func(id capability.ProtocolID) bool { return id == capability.ProtocolResponsesWebSocket })
		a.Credentials = maps.Clone(a.Credentials)
		if a.Credentials == nil {
			a.Credentials = map[string]any{}
		}
		a.Credentials[UpstreamProtocolsKey] = protocols
	}
	for _, key := range []string{"openai_oauth_responses_websockets_v2_mode", "openai_apikey_responses_websockets_v2_mode", "openai_oauth_responses_websockets_v2_enabled", "openai_apikey_responses_websockets_v2_enabled", "responses_websockets_v2_enabled", "openai_ws_enabled", "openai_ws_force_http"} {
		delete(extra, key)
	}
	extra[ResponsesWSConnectionModeKey] = mode
	a.Extra = extra
	return nil
}
