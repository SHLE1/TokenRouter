package app

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/imroc/req/v3"
)

// providePrivacyClientFactory 为隐私请求配置超时、Chrome 指纹和共享连接池。
func providePrivacyClientFactory() openai.PrivacyClientFactory {
	return func(proxyURL string) (*req.Client, error) {
		return httpclient.GetSharedReqClient(httpclient.ReqClientOptions{
			ProxyURL:    proxyURL,
			Timeout:     30 * time.Second,
			Impersonate: true,
		})
	}
}
