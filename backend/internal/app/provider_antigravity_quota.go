package app

import (
	"context"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/TokenFlux/TokenRouter/internal/egress"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideAntigravityQuota 为额度查询绑定模型响应读取上限和代理查询，代理缺失或读取失败时返回空地址与 false。
func provideAntigravityQuota(cfg *config.Config, proxies egress.ProxyRepository) *provider.AntigravityQuota {
	limit := resolveModelsListReadLimit(cfg)
	options := provideradapter.AntigravityQuotaOptions(limit, func(ctx context.Context, id int64) (string, bool) {
		if proxies == nil {
			return "", false
		}
		proxy, err := proxies.GetByID(ctx, id)
		if err != nil || proxy == nil {
			return "", false
		}
		return proxy.URL(), true
	})
	return &provider.AntigravityQuota{Options: options}
}
