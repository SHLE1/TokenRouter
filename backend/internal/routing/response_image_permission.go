package routing

// GroupAllowsResponsesImages 检查 Responses 图片策略，空分组默认允许，未设置策略时读取 AllowImageGeneration。
func GroupAllowsResponsesImages(group *Group) bool {
	return group == nil || group.ResponsesImagePolicy != "" || group.AllowImageGeneration
}
