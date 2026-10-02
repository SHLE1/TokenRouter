package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/TokenFlux/TokenRouter/internal/upstream/qoder"
)

// ModelRejectionDefaults 返回平台的默认模型，Qoder 站点无效时返回错误。
func ModelRejectionDefaults(platform string, site func() string) ([]string, error) {
	models := DefaultRequestModels(platform)
	if platform == capability.PlatformQoder {
		value, err := qoder.ParseSite(site())
		if err != nil {
			return nil, err
		}
		models = qoder.DefaultRequestModelIDsForSite(value)
	}
	return models, nil
}

type modelRejectionRules struct{ *provider.Record }

func (r modelRejectionRules) GetConfiguredRequestModels() []string {
	return r.Record.GetConfiguredRequestModels(provideradapter.ModelDefaults())
}

func (r modelRejectionRules) IsModelSupported(model string) bool {
	return r.Record.IsModelSupported(model, provideradapter.ModelDefaults(), provideradapter.ModelRules(r.Record))
}

// ModelRejectionProvider 为 routing 提供模型拒绝判断所需的提供商信息。
func ModelRejectionProvider(value *provider.Record) routing.ModelRejectionSource {
	return routing.ModelRejectionSource{Platform: value.Platform, Rules: modelRejectionRules{value}, Defaults: func(platform string) ([]string, error) {
		return ModelRejectionDefaults(platform, func() string { return value.GetCredential("site") })
	}}
}
