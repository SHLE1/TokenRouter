package provider

import (
	"github.com/TokenFlux/TokenRouter/internal/billing/pricing"
	"github.com/TokenFlux/TokenRouter/internal/protocol/wirejson"
	"github.com/TokenFlux/TokenRouter/internal/upstream"
	"github.com/TokenFlux/TokenRouter/internal/upstream/grok"
)

// GrokMediaCodec 绑定媒体协议处理所需的价格归一化函数。
// 输入归一化使用 billing/pricing 的规则。
func GrokMediaCodec() grok.MediaCodec {
	return grok.MediaCodec{Options: grok.MediaNormalization{
		MaxUploadPartSize:                             upstream.OpenAIImageMaxUploadPartSize,
		ImageTier1K:                                   pricing.ImageBillingSize1K,
		MarshalJSON:                                   wirejson.Marshal,
		NormalizeImageBillingTierOrDefault:            pricing.NormalizeImageBillingTierOrDefault,
		NormalizeVideoBillingResolutionOrDefault:      pricing.NormalizeVideoBillingResolutionOrDefault,
		NormalizeVideoBillingDurationSecondsOrDefault: pricing.NormalizeVideoBillingDurationSecondsOrDefault,
		ClassifyImageBillingTier:                      pricing.ClassifyImageBillingTier,
		ParseImageDimensions:                          pricing.ParseImageBillingDimensions,
	}}
}
