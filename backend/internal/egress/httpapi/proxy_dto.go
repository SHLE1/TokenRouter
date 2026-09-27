// 本文件维护 httpapi 的所属能力；兼容入口复用唯一实现。
package httpapi

import (
	proxydto "github.com/TokenFlux/TokenRouter/internal/egress/httpapi/dto"
)

type Proxy = proxydto.Proxy

type ProxyWithProviderCount = proxydto.ProxyWithProviderCount

type AdminProxy = proxydto.AdminProxy

type AdminProxyWithProviderCount = proxydto.AdminProxyWithProviderCount

type ProxyProviderSummary = proxydto.ProxyProviderSummary
