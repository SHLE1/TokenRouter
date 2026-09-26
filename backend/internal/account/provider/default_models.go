package provider

import (
	"slices"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/upstream/bedrock"

	"github.com/TokenFlux/TokenRouter/internal/account"
	"github.com/TokenFlux/TokenRouter/internal/upstream/anthropic"
	"github.com/TokenFlux/TokenRouter/internal/upstream/antigravity"
	"github.com/TokenFlux/TokenRouter/internal/upstream/deepseek"
	"github.com/TokenFlux/TokenRouter/internal/upstream/gemini"
	"github.com/TokenFlux/TokenRouter/internal/upstream/gemini/codeassist"
	"github.com/TokenFlux/TokenRouter/internal/upstream/grok"
	"github.com/TokenFlux/TokenRouter/internal/upstream/kimi"
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/TokenFlux/TokenRouter/internal/upstream/qoder"
	"github.com/TokenFlux/TokenRouter/internal/upstream/zhipu"
)

// DefaultAccountModels 复用各上游拥有的目录，未知平台没有隐式全模型能力。
func DefaultAccountModels(value *account.Record) []string {
	if value == nil {
		return nil
	}
	switch value.Platform {
	case account.PlatformAnthropic:
		if value.Type == account.AccountTypeBedrock {
			var models []string
			for alias, target := range bedrock.DefaultBedrockModelMapping {
				models = append(models, alias, target)
			}
			slices.Sort(models)
			return slices.Compact(models)
		}
		return anthropic.DefaultModelIDs()
	case account.PlatformOpenAI:
		if value.IsShadow() {
			return sparkModelVariants()
		}
		models := openai.DefaultModelIDs()
		if value.Type == account.AccountTypeAPIKey {
			models = append(models, "text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002")
		}
		return models
	case account.PlatformGemini:
		var models []string
		if value.IsGeminiGoogleOne() {
			for _, model := range codeassist.GoogleOneModels {
				models = append(models, model.ID)
			}
		} else if value.Type == account.AccountTypeOAuth {
			for _, model := range codeassist.DefaultModels {
				models = append(models, model.ID)
			}
		} else {
			for _, model := range gemini.DefaultModels() {
				models = append(models, strings.TrimPrefix(model.Name, "models/"))
			}
		}
		return models
	case account.PlatformAntigravity:
		var models []string
		for _, model := range antigravity.DefaultModels() {
			models = append(models, model.ID)
		}
		return models
	case account.PlatformGrok:
		return grok.DefaultModelIDs()
	case account.PlatformQoder:
		site, err := qoder.ParseSite(value.GetCredential("site"))
		if err != nil {
			return nil
		}
		return qoder.DefaultRequestModelIDsForSite(site)
	case account.PlatformKimi:
		return []string{kimi.DefaultTestModel}
	case account.PlatformZhipu:
		return []string{zhipu.DefaultTestModel}
	case account.PlatformDeepseek:
		return []string{deepseek.DefaultTestModel, "deepseek-reasoner"}
	default:
		return nil
	}
}
