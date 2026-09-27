// 本文件维护 httpapi 的所属能力；兼容入口复用唯一实现。
package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
	proxydto "github.com/TokenFlux/TokenRouter/internal/egress/httpapi/dto"
)

func ProxyFromService(p *egress.Proxy) *Proxy { return proxydto.ProxyFromEgress(p) }

func ProxyWithProviderCountFromService(p *egress.ProxyWithProviderCount) *ProxyWithProviderCount {
	return proxydto.ProxyWithProviderCountFromEgress(p)
}

func ProxyFromServiceAdmin(p *egress.Proxy) *AdminProxy { return proxydto.ProxyFromEgressAdmin(p) }

func ProxyWithProviderCountFromServiceAdmin(p *egress.ProxyWithProviderCount) *AdminProxyWithProviderCount {
	return proxydto.ProxyWithProviderCountFromEgressAdmin(p)
}

func ProxyProviderSummaryFromService(a *egress.ProxyProviderSummary) *ProxyProviderSummary {
	return proxydto.ProxyProviderSummaryFromEgress(a)
}
