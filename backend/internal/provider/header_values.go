package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/egress"
)

// HeaderOverrides 由提供商决定适用性，名称/值安全规则由 egress 唯一执行。
// 返回的 map 是 credentials 中 Header 配置的独立副本。
func (a *Record) HeaderOverrides() map[string]string {
	if !a.IsHeaderOverrideEnabled() {
		return nil
	}
	return egress.ResolveHeaderOverrides(StringMappingFromRaw(a.Credentials["header_overrides"]))
}
