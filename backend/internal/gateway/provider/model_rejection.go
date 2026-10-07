package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/provider"
	provideradapter "github.com/TokenFlux/TokenRouter/internal/provider/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing"
)

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
		return provideradapter.DefaultProviderModels(value), nil
	}}
}
