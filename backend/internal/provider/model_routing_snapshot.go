package provider

import (
	"maps"
)

// ModelRoutingSnapshot 保存本次模型匹配需要的规则。
// 在模型匹配时创建，并按需读取动态默认模型。
type ModelRoutingSnapshot struct {
	platform string
	mapping  map[string]string
}

func NewModelRoutingSnapshot(platform string, mapping map[string]string) ModelRoutingSnapshot {
	return ModelRoutingSnapshot{platform: platform, mapping: maps.Clone(mapping)}
}

func (s ModelRoutingSnapshot) Resolve(requested string) (string, bool) {
	return ResolveMappedModel(s.platform, s.mapping, requested)
}
