package forward

import (
	protocolanthropic "github.com/TokenFlux/TokenRouter/internal/protocol/anthropic"
	"github.com/TokenFlux/TokenRouter/internal/protocol/bridge"
	protocolopenai "github.com/TokenFlux/TokenRouter/internal/protocol/openai"
)

// AnthropicToResponses 为测试组合模型选项并调用协议转换。
func AnthropicToResponses(req *protocolanthropic.AnthropicRequest) (*protocolopenai.ResponsesRequest, error) {
	return bridge.AnthropicToResponses(req, ConversionOptionsForModel(req.Model))
}

func AnthropicToChatCompletionsRequest(req *protocolanthropic.AnthropicRequest) (*protocolopenai.ChatCompletionsRequest, error) {
	return bridge.AnthropicToChatCompletionsRequest(req, ConversionOptionsForModel(req.Model))
}

func ChatCompletionsToResponses(req *protocolopenai.ChatCompletionsRequest) (*protocolopenai.ResponsesRequest, error) {
	return bridge.ChatCompletionsToResponses(req, ConversionOptionsForModel(req.Model))
}
