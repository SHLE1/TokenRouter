package routing

import "strings"

// BillingModelForPrice 按价格配置的计费模型来源选择价格模型，且最早只从 Key 重定向目标开始。
func BillingModelForPrice(mapping GroupMappingResult, requestedModel, groupMappedModel, upstreamModel string) string {
	switch mapping.BillingModelSource {
	case BillingModelSourceRequested:
		return strings.TrimSpace(requestedModel)
	case BillingModelSourceUpstream:
		return strings.TrimSpace(upstreamModel)
	case BillingModelSourceGroupMapped:
		return strings.TrimSpace(groupMappedModel)
	default:
		return strings.TrimSpace(groupMappedModel)
	}
}
