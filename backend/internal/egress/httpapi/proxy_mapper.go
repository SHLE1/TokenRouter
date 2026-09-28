package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
	proxydto "github.com/TokenFlux/TokenRouter/internal/egress/httpapi/dto"
)

func ProxyFromServiceAdmin(p *egress.Proxy) *AdminProxy { return proxydto.ProxyFromEgressAdmin(p) }

func ProxyWithProviderCountFromServiceAdmin(p *egress.ProxyWithProviderCount) *AdminProxyWithProviderCount {
	return proxydto.ProxyWithProviderCountFromEgressAdmin(p)
}

func ProxyProviderSummaryFromService(a *egress.ProxyProviderSummary) *ProxyProviderSummary {
	return proxydto.ProxyProviderSummaryFromEgress(a)
}
