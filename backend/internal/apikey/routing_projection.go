package apikey

import "github.com/TokenFlux/TokenRouter/internal/routing"

// GroupFromRouting 复制分组，供认证与管理调用方各自使用。
func GroupFromRouting(group *routing.Group) *routing.Group {
	return routing.CloneGroup(group)
}

// RoutingGroup 返回请求独立副本，调用方不能修改认证缓存。
func RoutingGroup(group *routing.Group) *routing.Group {
	return routing.CloneGroup(group)
}
