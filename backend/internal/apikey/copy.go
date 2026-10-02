package apikey

import (
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

// CopyAPIKey 复制 Key 及关联的用户和分组数据。认证缓存使用自身快照的复制规则。
func CopyAPIKey(value *APIKey) *APIKey {
	if value == nil {
		return nil
	}
	out := *value
	out.User = identity.CopyUser(value.User)
	out.ActorUser = identity.CopyUser(value.ActorUser)
	out.Group = routing.CloneGroup(value.Group)
	out.CompositeGroups = CopyCompositeGroups(value.CompositeGroups)
	return &out
}

// CopyCompositeGroup 保留绑定顺序和请求级覆盖，分组策略不与输入共享。
func CopyCompositeGroup(value APIKeyCompositeGroup) APIKeyCompositeGroup {
	value.Group = routing.CloneGroup(value.Group)
	return value
}

func CopyCompositeGroups(values []APIKeyCompositeGroup) []APIKeyCompositeGroup {
	if values == nil {
		return nil
	}
	out := make([]APIKeyCompositeGroup, len(values))
	for i, value := range values {
		out[i] = CopyCompositeGroup(value)
	}
	return out
}
