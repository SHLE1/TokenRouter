package app

import (
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideOpenAIImages 绑定共享的请求、输出和任务跟踪器，模型冷却由提供商模块管理。
func provideOpenAIImages(text *gatewayhttp.OpenAITextExecutor, activity *gatewayRequestActivity) *gatewayhttp.OpenAIImagesExecutor {
	cooldown := &provideradapter.ImageToolCooldown{Store: text.Requests.Providers}
	if text.Requests.Readers != nil {
		cooldown.Settings = text.Requests.Readers.Provider.GetOpenAIImagesOAuthUnavailableCooldownSettings
	}
	result := &gatewayhttp.OpenAIImagesExecutor{Requests: text.Requests, Output: text.Output, Cooldown: cooldown, Enter: activity.Enter}
	return result
}
