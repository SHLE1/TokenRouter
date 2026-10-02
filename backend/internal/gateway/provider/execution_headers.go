package provider

import (
	"net/http"

	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// BindExecutionHeaders 返回绑定提供商的 Header 设置函数，调用时读取最新字段。
func BindExecutionHeaders(value *ExecutionProvider) func(http.Header) {
	return func(headers http.Header) {
		provideradapter.ApplyProviderHeaderOverrides(ExecutionProtocolRecord(value), headers)
	}
}

// BindExecutionHeaderValue 返回绑定提供商的 Header 查询函数，调用时读取字段值。
func BindExecutionHeaderValue(value *ExecutionProvider) func(string) (string, bool) {
	return func(name string) (string, bool) {
		return provideradapter.HeaderOverrideValue(ExecutionProtocolRecord(value), name)
	}
}
