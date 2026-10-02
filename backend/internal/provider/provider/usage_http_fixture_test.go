package provider

import (
	"net/http"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/infra/httpclient/tlsfingerprint"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// newUsageContractService 为 HTTP 测试构造用量查询实例，传输替身记录请求和取消。
func newUsageContractService(reader provider.UpstreamUsageReader, transport interface {
	DoWithTLS(*http.Request, string, int64, int, *tlsfingerprint.Profile) (*http.Response, error)
}, policy egress.UsageURLPolicy,
) *provider.UpstreamUsageService {
	options := UsageHTTPOptions{Available: reader != nil && transport != nil, Policy: policy}
	if transport != nil {
		options.Do = transport.DoWithTLS
	}
	return provider.NewUpstreamUsageService(reader, NewUpstreamUsageHTTPExecution(options), provider.UpstreamUsageOptions{Now: time.Now})
}
