package app

import (
	"strings"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/config"
	gatewayhttp "github.com/TokenFlux/TokenRouter/internal/gateway/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// provideOpenAIHTTPResources 构造供文本、媒体、WS 和兼容入口共享的 HTTP 资源。
func provideOpenAIHTTPResources(concurrency *scheduler.ConcurrencyService, cfg *config.Config) *gatewayhttp.OpenAIHTTPResources {
	ping := time.Duration(0)
	imageOptions := openAIImageAdmissionOptions(cfg)
	if cfg != nil {
		ping = time.Duration(cfg.Concurrency.PingInterval) * time.Second
	}
	return &gatewayhttp.OpenAIHTTPResources{Concurrency: gatewayhttp.NewConcurrencyHelper(concurrency, gatewayhttp.SSEPingFormatComment, ping), Images: &scheduler.ImageConcurrencyLimiter{}, ImageOptions: imageOptions}
}

// openAIImageAdmissionOptions 返回图片准入所需的配置参数。
func openAIImageAdmissionOptions(cfg *config.Config) *gatewayhttp.OpenAIImageAdmissionOptions {
	if cfg == nil {
		return nil
	}
	value := cfg.Gateway.ImageConcurrency
	return &gatewayhttp.OpenAIImageAdmissionOptions{Enabled: value.Enabled, Limit: value.MaxConcurrentRequests, Wait: strings.TrimSpace(value.OverflowMode) == config.ImageConcurrencyOverflowModeWait, Timeout: time.Duration(value.WaitTimeoutSeconds) * time.Second, MaxWaiting: value.MaxWaitingRequests}
}
