package httpapi

import (
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

// OpenAIResponsesExecutor 组合固定请求、输出与协议执行能力；连接资源由各传输拥有者管理。
type OpenAIResponsesExecutor struct {
	Requests    *OpenAIRequests
	Output      *OpenAIResponseOutput
	Text        *OpenAITextExecutor
	Grok        *GrokExecutor
	Lineage     *OpenAIEncryptedLineage
	ImageBridge *gatewayadapter.ResponseImagePolicy
}
