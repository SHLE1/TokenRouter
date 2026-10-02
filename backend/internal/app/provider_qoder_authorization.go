package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
)

// provideQoderAuthorization 绑定代理读取接口，授权组件和供应商执行器分别管理会话与请求。
func provideQoderAuthorization(proxies egress.ProxyRepository) *provideradapter.QoderAuthorization {
	return provideradapter.NewQoderAuthorization(func(ctx context.Context, id *int64) (string, error) {
		if id == nil {
			return "", nil
		}
		if proxies == nil {
			return "", errors.New("proxy repository is not configured")
		}
		proxy, err := proxies.GetByID(ctx, *id)
		if err != nil {
			return "", fmt.Errorf("get proxy: %w", err)
		}
		if proxy == nil {
			return "", errors.New("proxy not found")
		}
		return proxy.URL(), nil
	}, nil)
}
